package main

import (
	"encoding/json"
	"fmt"
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
			return nil, fmt.Errorf("config path: %w", err)
		}
		return &Config{path: path}, nil
	}
	cfg := &Config{path: path}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func (cfg *Config) Save() error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.path), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(cfg.path, b, 0644); err != nil {
		return fmt.Errorf("write config %s: %w", cfg.path, err)
	}
	return nil
}
