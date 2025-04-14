package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Core struct {
	} `mapstructure:"core"`

	Plugins struct {
		Path string   `mapstructure:"path"`
		Load []string `mapstructure:"load"`
	} `mapstructure:"plugins"`

	Log struct {
		Level string `mapstructure:"level"`
		File  string `mapstructure:"file"`
	} `mapstructure:"log"`
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

	return nil
}
