// nacos/nacos.go
package nacos

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CheckAuth 通过 Nacos HTTP 登录接口验证指定节点的连通性及认证是否成功。
// address 格式为 "host:port"（仅 HTTP 端口，通常为 8848）。
// 成功返回 nil，失败返回描述错误的非 nil 错误。
func CheckAuth(address, username, password string, timeout time.Duration) error {
	loginURL := fmt.Sprintf("http://%s/nacos/v1/auth/login", address)

	body := url.Values{}
	body.Set("username", username)
	body.Set("password", password)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(loginURL, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()
	// 读取并丢弃响应体，确保连接能被复用
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("认证失败，状态码: %d", resp.StatusCode)
	}

	return nil
}
