// checker/zookeeper_test.go
package checker

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// startMockZookeeper 启动一个模拟 Zookeeper 服务器，响应 "ruok" 命令返回 "imok"。
// 返回监听地址和停止函数。
func startMockZookeeper(t *testing.T, response string) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("启动模拟服务器失败: %v", err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 4)
				if _, err := c.Read(buf); err != nil {
					return
				}
				if string(buf) == "ruok" {
					c.Write([]byte(response))
				}
			}(conn)
		}
	}()

	return ln.Addr().String(), func() { ln.Close() }
}

func TestCheckZookeeperNode_Success(t *testing.T) {
	addr, stop := startMockZookeeper(t, "imok")
	defer stop()

	err := checkZookeeperNode(addr, 3*time.Second)
	if err != nil {
		t.Errorf("期望连接成功，但得到错误: %v", err)
	}
}

func TestCheckZookeeperNode_UnexpectedResponse(t *testing.T) {
	addr, stop := startMockZookeeper(t, "fail")
	defer stop()

	err := checkZookeeperNode(addr, 3*time.Second)
	if err == nil {
		t.Error("期望收到错误（意外响应），但得到 nil")
	}
}

func TestCheckZookeeperNode_ConnectionRefused(t *testing.T) {
	// 使用一个不存在的地址
	err := checkZookeeperNode("127.0.0.1:1", 500*time.Millisecond)
	if err == nil {
		t.Error("期望连接被拒绝，但得到 nil")
	}
}

func TestCheckZookeeperConnection_SingleHost(t *testing.T) {
	addr, stop := startMockZookeeper(t, "imok")
	defer stop()

	err := CheckZookeeperConnection(addr, 3*time.Second)
	if err != nil {
		t.Errorf("期望连接成功，但得到错误: %v", err)
	}
}

func TestCheckZookeeperConnection_ConnectString_OneHealthy(t *testing.T) {
	// 只有第二个节点可用
	addr, stop := startMockZookeeper(t, "imok")
	defer stop()

	connectString := fmt.Sprintf("127.0.0.1:1,%s", addr)
	err := CheckZookeeperConnection(connectString, 3*time.Second)
	if err != nil {
		t.Errorf("期望至少一个节点可达，但得到错误: %v", err)
	}
}

func TestCheckZookeeperConnection_ConnectString_AllFail(t *testing.T) {
	connectString := "127.0.0.1:1,127.0.0.1:2"
	err := CheckZookeeperConnection(connectString, 500*time.Millisecond)
	if err == nil {
		t.Error("期望所有节点不可达时返回错误，但得到 nil")
	}
}

func TestCheckZookeeperConnection_WithChroot(t *testing.T) {
	addr, stop := startMockZookeeper(t, "imok")
	defer stop()

	connectString := addr + "/mychroot"
	err := CheckZookeeperConnection(connectString, 3*time.Second)
	if err != nil {
		t.Errorf("期望忽略 chroot 路径后连接成功，但得到错误: %v", err)
	}
}

func TestCheckZookeeperConnection_Empty(t *testing.T) {
	err := CheckZookeeperConnection("", 3*time.Second)
	if err == nil {
		t.Error("期望空连接字符串返回错误，但得到 nil")
	}
}

func TestParseZookeeperConnectString(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "host1:2181",
			expected: []string{"host1:2181"},
		},
		{
			input:    "host1:2181,host2:2181,host3:2181",
			expected: []string{"host1:2181", "host2:2181", "host3:2181"},
		},
		{
			input:    "host1:2181,host2:2181/chroot",
			expected: []string{"host1:2181", "host2:2181"},
		},
		{
			input:    " host1:2181 , host2:2181 ",
			expected: []string{"host1:2181", "host2:2181"},
		},
		{
			input:    "",
			expected: []string{},
		},
	}

	for _, tc := range tests {
		got := parseZookeeperConnectString(tc.input)
		if len(got) != len(tc.expected) {
			t.Errorf("输入 %q: 期望 %v，得到 %v", tc.input, tc.expected, got)
			continue
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("输入 %q: 第 %d 项期望 %q，得到 %q", tc.input, i, tc.expected[i], got[i])
			}
		}
	}
}
