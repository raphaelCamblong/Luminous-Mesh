package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type TLSConfig struct {
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
	CACert   string `toml:"ca_cert"`
}

type AuthConfig struct {
	TokenSecret   string        `toml:"token_secret"`
	TokenDuration time.Duration `toml:"token_duration"`
	CACertPath    string        `toml:"ca_cert_path"`
	CAKeyPath     string        `toml:"ca_key_path"`
}

type CoreConfig struct {
	ListenAddr       string            `toml:"listen_addr"`
	APIEndpoint      string            `toml:"api_endpoint"`
	TLS              TLSConfig         `toml:"tls"`
	Auth             AuthConfig        `toml:"auth"`
	ConnectionParams map[string]string `toml:"connection_params"`
}

type PluginsConfig struct {
	Path string   `toml:"path"`
	Load []string `toml:"load"`
}

type LogConfig struct {
	Level string `toml:"level"`
	File  string `toml:"file"`
}

type Config struct {
	Core    CoreConfig    `toml:"core"`
	Plugins PluginsConfig `toml:"plugins"`
	Log     LogConfig     `toml:"log"`
}

func DefaultConfig() *Config {
	return &Config{
		Core: CoreConfig{
			ListenAddr:  ":50051",
			APIEndpoint: "localhost:50051",
			TLS: TLSConfig{
				CertFile: "/etc/luminous-mesh/certs/server.crt",
				KeyFile:  "/etc/luminous-mesh/certs/server.key",
			},
			Auth: AuthConfig{
				TokenSecret:   "default-secret",
				TokenDuration: 24 * time.Hour,
				CACertPath:    "/etc/luminous-mesh/certs/ca.crt",
				CAKeyPath:     "/etc/luminous-mesh/certs/ca.key",
			},
			ConnectionParams: map[string]string{
				"max_reconnect_delay": "60s",
				"keepalive_time":      "30s",
			},
		},
		Plugins: PluginsConfig{
			Path: "/etc/luminous-mesh/plugins",
			Load: []string{},
		},
		Log: LogConfig{
			Level: "info",
			File:  "/var/log/luminous-mesh/app.log",
		},
	}
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		args := ParseArgs()

		cfg, err := LoadConfig(args.ConfigPath)
		if err != nil {
			panic(fmt.Errorf("❌ config load failed: %w", err))
		}
		err = cfg.Validate()
		if err != nil {
			panic(fmt.Errorf("❌ config validation failed: %w", err))
		}
		instance = cfg
	})
	return instance
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to open config file: %w", err)
	}
	defer f.Close()

	if _, err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("❌ Failed to decode config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Plugins.Path == "" {
		return fmt.Errorf("❌ Plugins path is not set")
	}

	if len(c.Plugins.Load) == 0 {
		return fmt.Errorf("❌ No plugins to load")
	}

	if c.Core.ListenAddr == "" {
		return fmt.Errorf("listen_addr is required")
	}

	if c.Core.APIEndpoint == "" {
		return fmt.Errorf("api_endpoint is required")
	}

	if err := validateTLSConfig(&c.Core.TLS); err != nil {
		return fmt.Errorf("invalid TLS configuration: %w", err)
	}

	if err := validateAuthConfig(&c.Core.Auth); err != nil {
		return fmt.Errorf("invalid auth configuration: %w", err)
	}

	if err := validateConnectionParams(&c.Core.ConnectionParams); err != nil {
		return fmt.Errorf("invalid connection parameters: %w", err)
	}
	return nil
}

func validateTLSConfig(config *TLSConfig) error {
	if config.CertFile == "" {
		return fmt.Errorf("cert_file is required")
	}

	if config.KeyFile == "" {
		return fmt.Errorf("key_file is required")
	}

	if len(config.CACert) == 0 {
		return fmt.Errorf("ca_cert is required")
	}

	return nil
}

func validateAuthConfig(config *AuthConfig) error {
	if config.TokenSecret == "" {
		return fmt.Errorf("token_secret is required")
	}

	if config.TokenDuration == 0 {
		return fmt.Errorf("token_duration is required")
	}

	if config.CACertPath == "" {
		return fmt.Errorf("ca_cert_path is required")
	}

	if config.CAKeyPath == "" {
		return fmt.Errorf("ca_key_path is required")
	}

	return nil
}

func validateConnectionParams(config *map[string]string) error {
	if _, ok := (*config)["max_reconnect_delay"]; !ok {
		return fmt.Errorf("max_reconnect_delay is required")
	}

	if _, ok := (*config)["keepalive_time"]; !ok {
		return fmt.Errorf("keepalive_time is required")
	}

	return nil
}
