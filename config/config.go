// config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// NetworkConfig 定义了各个服务的地址列表，键为服务名称，值为字符串数组
type NetworkConfig map[string][]string

// KibanaConfig 定义了 Kibana 服务的连接配置，支持用户名和密码认证
type KibanaConfig struct {
	Addresses []string `yaml:"addresses"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

// MinioConfig 保存 MinIO 的认证信息及地址列表
type MinioConfig struct {
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	UseSSL    bool     `yaml:"use_ssl"`
	Addresses []string `yaml:"addresses"`
}

// RedisConfig holds Redis-specific configuration including an optional password
// and the list of Redis addresses to check.
type RedisConfig struct {
	Password  string   `yaml:"password"`
	Addresses []string `yaml:"addresses"`
}

// AppConfig 顶层配置结构，包含通用 TCP 服务、MinIO、Kibana 和 Redis 专项配置
type AppConfig struct {
	Services NetworkConfig `yaml:"services"`
	Minio    *MinioConfig  `yaml:"minio"`
	Kibana   KibanaConfig  `yaml:"kibana"`
	Redis    RedisConfig   `yaml:"redis"`
}

// LoadConfig 从指定文件中读取 YAML 配置，并解析为 AppConfig 类型
func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
