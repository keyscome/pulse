// checker/checker.go
package checker

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// CheckConnection 尝试在指定超时时间内建立 TCP 连接
// 成功返回 nil，失败返回错误信息
func CheckConnection(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// CheckKibanaConnection 通过 HTTP 请求 Kibana 状态接口，支持用户名和密码认证
// 成功返回 nil，失败返回错误信息
func CheckKibanaConnection(address, username, password string, timeout time.Duration) error {
	url := fmt.Sprintf("http://%s/api/status", address)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}
	return nil
}
