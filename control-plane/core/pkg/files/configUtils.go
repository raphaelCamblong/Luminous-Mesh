package files

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func FindConfigPath() string {
	var configPath string
	flag.StringVar(&configPath, "c", "", "path to config file")
	flag.Parse()

	if configPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		configPath = filepath.Join(wd, "config.toml")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Errorf("❌ config file not found at %s", configPath))
	}

	return configPath
}
