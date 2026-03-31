// checker/checker.go
package checker

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// CheckConnection attempts to establish a TCP connection within the specified timeout.
// Returns nil on success, or an error on failure.
func CheckConnection(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// CheckRedisConnection connects to a Redis instance at address, optionally
// authenticates with password (when non-empty), and verifies the connection
// by sending a PING command. Returns nil on success.
func CheckRedisConnection(address, password string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("设置连接超时失败: %v", err)
	}

	reader := bufio.NewReader(conn)

	if password != "" {
		// Send: AUTH <password>
		cmd := fmt.Sprintf("*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(password), password)
		if _, err = fmt.Fprint(conn, cmd); err != nil {
			return fmt.Errorf("发送 AUTH 命令失败: %v", err)
		}
		resp, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("读取 AUTH 响应失败: %v", err)
		}
		if resp = strings.TrimSpace(resp); resp != "+OK" {
			return fmt.Errorf("Redis AUTH 失败: %s", resp)
		}
	}

	// Send: PING
	if _, err = fmt.Fprint(conn, "*1\r\n$4\r\nPING\r\n"); err != nil {
		return fmt.Errorf("发送 PING 命令失败: %v", err)
	}
	resp, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取 PING 响应失败: %v", err)
	}
	if resp = strings.TrimSpace(resp); resp != "+PONG" {
		return fmt.Errorf("Redis PING 失败: %s", resp)
	}

	return nil
}
