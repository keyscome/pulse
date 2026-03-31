// config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// NetworkConfig 定义了各个服务的地址列表，键为服务名称，值为字符串数组
type NetworkConfig map[string][]string

// MinioConfig 保存 MinIO 的认证信息及地址列表
type MinioConfig struct {
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	UseSSL    bool     `yaml:"use_ssl"`
	Addresses []string `yaml:"addresses"`
}

// AppConfig 顶层配置结构，包含通用 TCP 服务和 MinIO 专项配置
type AppConfig struct {
	Services NetworkConfig `yaml:"services"`
	Minio    *MinioConfig  `yaml:"minio"`
}

// LoadConfig 从指定文件中读取 YAML 配置，并解析为 AppConfig 类型
func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
