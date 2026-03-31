// checker/elasticsearch.go
package checker

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// CheckElasticsearch 通过 Elasticsearch HTTP API（/_cluster/health）验证单个节点的连通性。
// 若提供了 username 或 password，则使用 HTTP Basic Auth 进行认证。
func CheckElasticsearch(address, username, password string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	url := fmt.Sprintf("http://%s/_cluster/health", address)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()
	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 状态 %d", resp.StatusCode)
	}
	return nil
}
