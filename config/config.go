// config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// KibanaConfig 定义了 Kibana 服务的连接配置，支持用户名和密码认证
type KibanaConfig struct {
	Addresses []string `yaml:"addresses"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

// AppConfig 定义了整体配置结构，Kibana 使用独立的认证配置，其余服务使用地址列表
type AppConfig struct {
	Kibana   KibanaConfig        `yaml:"kibana"`
	Services map[string][]string `yaml:",inline"`
}

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

// AppConfig 顶层配置结构，包含通用 TCP 服务、MinIO 和 Kibana 专项配置
type AppConfig struct {
	Services NetworkConfig `yaml:"services"`
	Minio    *MinioConfig  `yaml:"minio"`
	Kibana   KibanaConfig  `yaml:"kibana"`
}

// LoadConfig 从指定文件中读取 YAML 配置，并解析为 AppConfig 类型
func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}
	var cfg AppConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return AppConfig{}, err
	}
	return &cfg, nil
}
