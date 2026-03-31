// config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// NetworkConfig 定义了各个服务的地址列表，键为服务名称，值为字符串数组
type NetworkConfig map[string][]string

// NacosConfig 保存 Nacos 集群连接配置，包括节点地址列表和认证凭据
type NacosConfig struct {
	Addresses []string `yaml:"addresses"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

// Config 保存所有服务的连接配置：
//   - Network：通用 TCP 检测服务（内联到顶层键）
//   - Nacos：带认证的 Nacos 集群专属配置
type Config struct {
	Network NetworkConfig `yaml:",inline"`
	Nacos   *NacosConfig  `yaml:"nacos"`
}

// LoadConfig 从指定文件中读取 YAML 配置，并解析为 Config 类型
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
