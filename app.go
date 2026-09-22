package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

type App struct {
	ApiKey     string
	Depotcache string
	Out        *Output
	Config     *Config

	SLSconfigPath    string
	SlsConfig        *yaml.Node // Root node of config.yaml
	AdditionalApps   *yaml.Node // []appid
	AdditionalDepots *yaml.Node // []depotid
	DecryptionKeys   *yaml.Node // map[depotid]key
	ManifestIds      *yaml.Node // map[depotid]gid
	AppTokens        *yaml.Node // map[appid]token
}

func NewApp(apikey string, verbose bool, loadcfg bool, slsc string) (*App, error) {
	depotcache, err := xdg.SearchDataFile("Steam/depotcache")
	if err != nil {
		return nil, errors.New("Steam depotcache not found, start any game's installation in Steam.")
	}
	if slsc == "" {
		slsc, err = xdg.SearchConfigFile("SLSsteam/config.yaml")
		if err != nil {
			return nil, errors.New("SLSsteam config not found, install and run once.")
		}
	}

	b, err := os.ReadFile(slsc)
	if err != nil {
		return nil, fmt.Errorf("Failed to read SLSsteam config: %w", err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("SLSsteam config is invalid: %w", err)
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("SLSsteam config is empty")
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("Invalid SLSsteam config format in %q: expected top-level map, got %v", slsc, root.Kind)
	}

	a := &App{
		Depotcache:    depotcache,
		SLSconfigPath: slsc,
		SlsConfig:     root,
		ApiKey:        apikey,
	}

	for _, f := range []struct {
		key    string
		target **yaml.Node
		kind   yaml.Kind
		tag    string
	}{
		{"AdditionalApps", &a.AdditionalApps, yaml.SequenceNode, "!!seq"},
		{"AdditionalDepots", &a.AdditionalDepots, yaml.SequenceNode, "!!seq"},
		{"DecryptionKeys", &a.DecryptionKeys, yaml.MappingNode, "!!map"},
		{"ManifestIds", &a.ManifestIds, yaml.MappingNode, "!!map"},
		{"AppTokens", &a.AppTokens, yaml.MappingNode, "!!map"},
	} {
		res, err := EnsureKey(root, f.key, f.kind, f.tag)
		if err != nil {
			return nil, fmt.Errorf("SLSsteam/config.yaml: %w", err)
		}
		*f.target = res
	}
	if loadcfg {
		a.Config, err = LoadConfig()
		if err != nil {
			return nil, err
		}
	}
	a.Out = NewOutput(os.Stdout, verbose)
	return a, nil
}

func (a *App) SaveSLSconfig() error {
	b, err := yaml.Marshal(a.SlsConfig)
	if err != nil {
		return fmt.Errorf("SLSsteam/config.yaml: %w", err)
	}
	if err := os.WriteFile(a.SLSconfigPath, b, 0644); err != nil {
		return fmt.Errorf("SLSsteam/config.yaml: %w", err)
	}
	return nil
}
