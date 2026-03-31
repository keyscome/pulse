package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return filepath.Clean(f.Name())
}

func TestLoadConfig_RedisWithPassword(t *testing.T) {
	yml := `
redis:
  password: "s3cret"
  addresses:
    - 127.0.0.1:6379
    - 127.0.0.1:6380
nacos:
  - 127.0.0.1:8848
`
	path := writeTemp(t, yml)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Redis.Password != "s3cret" {
		t.Errorf("Redis.Password = %q, want %q", cfg.Redis.Password, "s3cret")
	}
	if len(cfg.Redis.Addresses) != 2 {
		t.Errorf("Redis.Addresses length = %d, want 2", len(cfg.Redis.Addresses))
	}
	if cfg.Redis.Addresses[0] != "127.0.0.1:6379" {
		t.Errorf("Redis.Addresses[0] = %q, want %q", cfg.Redis.Addresses[0], "127.0.0.1:6379")
	}

	nacosAddrs, ok := cfg.Services["nacos"]
	if !ok || len(nacosAddrs) != 1 || nacosAddrs[0] != "127.0.0.1:8848" {
		t.Errorf("Services[nacos] = %v, want [127.0.0.1:8848]", nacosAddrs)
	}
}

func TestLoadConfig_RedisNoPassword(t *testing.T) {
	yml := `
redis:
  addresses:
    - 127.0.0.1:6379
`
	path := writeTemp(t, yml)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Redis.Password != "" {
		t.Errorf("Redis.Password = %q, want empty", cfg.Redis.Password)
	}
	if len(cfg.Redis.Addresses) != 1 {
		t.Errorf("Redis.Addresses length = %d, want 1", len(cfg.Redis.Addresses))
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
