// config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// RedisConfig holds Redis-specific configuration including an optional password
// and the list of Redis addresses to check.
type RedisConfig struct {
	Password  string   `yaml:"password"`
	Addresses []string `yaml:"addresses"`
}

// Config holds the full application configuration.
// Redis is parsed as a structured block; all other services are treated as
// plain address lists and collected in Services.
type Config struct {
	Redis    RedisConfig         `yaml:"redis"`
	Services map[string][]string `yaml:",inline"`
}

// LoadConfig reads the YAML configuration file at path and returns a Config.
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
