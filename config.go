package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

type Config struct {
	Games  []int  `json:"managed-games"`
	Pins   []int  `json:"pins"`
	ApiKey string `json:"api-key"`
	path   string
}

func LoadConfig() (*Config, error) {
	path, err := xdg.SearchConfigFile("huhcap/config.json")
	if err != nil {
		path, err = xdg.ConfigFile("huhcap/config.json")
		if err != nil {
			return nil, err
		}
		return &Config{path: path}, nil
	}
	cfg := &Config{path: path}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *Config) Save() error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.path), 0755); err != nil {
		return err
	}
	return os.WriteFile(cfg.path, b, 0644)
}
