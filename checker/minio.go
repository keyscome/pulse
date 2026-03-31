// checker/minio.go
package checker

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// CheckMinioConnection 使用用户名和密码对 MinIO 服务进行认证连接检测。
// 成功返回 nil，失败返回错误信息。
func CheckMinioConnection(address, username, password string, useSSL bool, timeout time.Duration) error {
	client, err := minio.New(address, &minio.Options{
		Creds:  credentials.NewStaticV4(username, password, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("创建 MinIO 客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("MinIO 认证连接失败: %w", err)
	}

	return nil
}
