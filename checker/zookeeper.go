// checker/zookeeper.go
package checker

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// CheckZookeeperConnection 验证 Zookeeper 节点的连通性，使用 Zookeeper 四字命令 "ruok"。
// connectString 可以是单个 "host:port"，也可以是标准 Zookeeper 连接字符串，
// 例如 "host1:2181,host2:2181,host3:2181" 或 "host1:2181,host2:2181/chroot"。
// 只要有至少一个节点响应 "imok" 即返回 nil。
func CheckZookeeperConnection(connectString string, timeout time.Duration) error {
	hosts := parseZookeeperConnectString(connectString)
	if len(hosts) == 0 {
		return fmt.Errorf("Zookeeper 连接字符串为空")
	}

	var errs []string
	for _, host := range hosts {
		if err := checkZookeeperNode(host, timeout); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v", host, err))
		}
	}
	return fmt.Errorf("所有 Zookeeper 节点均不可达: %s", strings.Join(errs, "; "))
}

// parseZookeeperConnectString 将 Zookeeper 连接字符串解析为单独的 host:port 列表。
// 支持可选的 chroot 后缀，例如 "host1:2181,host2:2181/chroot"。
func parseZookeeperConnectString(connectString string) []string {
	// 去除可选的 chroot 路径（如 "/chroot"），仅当斜杠前包含冒号时才视为 chroot
	if idx := strings.Index(connectString, "/"); idx != -1 {
		before := connectString[:idx]
		if strings.Contains(before, ":") {
			connectString = before
		}
	}

	parts := strings.Split(connectString, ",")
	hosts := make([]string, 0, len(parts))
	for _, part := range parts {
		host := strings.TrimSpace(part)
		if host != "" {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

// checkZookeeperNode 向单个 Zookeeper 节点发送 "ruok" 四字命令，
// 并验证其是否回应 "imok"。
func checkZookeeperNode(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}

	if _, err := conn.Write([]byte("ruok")); err != nil {
		return err
	}

	buf := make([]byte, 4)
	if _, err := conn.Read(buf); err != nil {
		return err
	}

	if string(buf) != "imok" {
		return fmt.Errorf("Zookeeper 节点 %s 返回意外响应: %q", address, string(buf))
	}
	return nil
}
