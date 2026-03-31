// config/config.go
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// NetworkConfig 定义了各个服务的地址列表，键为服务名称，值为字符串数组
type NetworkConfig map[string][]string

// ElasticsearchConfig 保存 Elasticsearch 集群的连接参数
type ElasticsearchConfig struct {
	Addresses []string `yaml:"addresses"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

// AppConfig 是完整的应用配置，包含结构化的 Elasticsearch 配置和其他 TCP 服务的地址列表
type AppConfig struct {
	Elasticsearch *ElasticsearchConfig
	Network       NetworkConfig
}

// rawConfig 用于从 YAML 文件中单独解析 elasticsearch 配置节
type rawConfig struct {
	Elasticsearch *ElasticsearchConfig `yaml:"elasticsearch"`
}

// LoadConfig 从指定文件中读取 YAML 配置。
// elasticsearch 键解析为 ElasticsearchConfig（支持用户名/密码认证）；
// 其他所有键视为 TCP 地址列表。
func LoadConfig(path string) (*AppConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 解析结构化的 elasticsearch 节
	var raw rawConfig
	if err = yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// 将所有键解析为通用 map，以便提取 TCP 服务地址列表
	var all map[string]interface{}
	if err = yaml.Unmarshal(data, &all); err != nil {
		return nil, err
	}

	network := make(NetworkConfig)
	for k, v := range all {
		if k == "elasticsearch" {
			continue // 由 ElasticsearchConfig 单独处理
		}
		if list, ok := v.([]interface{}); ok {
			addrs := make([]string, 0, len(list))
			for _, a := range list {
				if s, ok := a.(string); ok {
					addrs = append(addrs, s)
				}
			}
			network[k] = addrs
		}
	}

	return &AppConfig{
		Elasticsearch: raw.Elasticsearch,
		Network:       network,
	}, nil
}
