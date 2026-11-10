package main

import (
	"fmt"
	"slices"

	"github.com/alecthomas/kong"
)

type InstallCmd struct {
	AppIDs []string `arg:"" required:"" placeholder:"APPID" help:"Games to install; can be a appid, store page or game name to search."`
}

func (c *InstallCmd) Run(cli *CLI) error {
	a, err := NewApp(cli.ApiKey, cli.Verbose, cli.SLSconfig)
	if err != nil {
		return err
	}
	var found []foundGame
	for _, g := range c.AppIDs {
		appid, name, err := a.parseAppID(g)
		if err != nil {
			return fmt.Errorf("parse %q: %w", g, err)
		}
		if name == "" {
			name, err = getGameName(appid)
			if err != nil {
				return fmt.Errorf("get name for %d: %w", appid, err)
			}
		}
		found = append(found, foundGame{appid, name})
	}

	// TODO add limit check
	var upd bool
	i := 0
	for _, g := range found {
		if err := a.InstallGame(g.AppID, g.Name); err != nil {
			err = fmt.Errorf("install %s (%d): %w", g.Name, g.AppID, err)
			if upd {
				if saveErr := a.Config.Save(); saveErr != nil {
					err = fmt.Errorf("%w; also failed to save config: %v", err, saveErr)
				}
			}
			if i > 0 {
				if saveErr := a.SaveSLSconfig(); saveErr != nil {
					err = fmt.Errorf("%w; %w", err, saveErr)
				}
			}
			return err
		}
		if _, ok := a.Config.Games[g.AppID]; !ok {
			a.Config.Games[g.AppID] = g.Name
			upd = true
		}
		i++
		a.Out.Info("Installed %s!", g.Name)
	}
	if upd {
		if err := a.Config.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
	}
	return a.SaveSLSconfig()
}

type UpdateCmd struct {
	DryRun bool `help:"Just show what games are outdated."`
}

func (c *UpdateCmd) Run(cli *CLI) error {
	a, err := NewApp(cli.ApiKey, cli.Verbose, cli.SLSconfig)
	if err != nil {
		return err
	}
	return a.Update(c.DryRun)
}

type PinCmd struct {
	AppID int `arg:"" placeholder:"APPID" help:"AppID to pin."`
}

func (c *PinCmd) Run(cli *CLI) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if !slices.Contains(cfg.Pins, c.AppID) {
		cfg.Pins = append(cfg.Pins, c.AppID)
		return cfg.Save()
	}
	return nil
}

type UnpinCmd struct {
	AppID int `arg:"" placeholder:"APPID" help:"AppID to unpin."`
}

func (c *UnpinCmd) Run(cli *CLI) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	old := len(cfg.Pins)
	cfg.Pins = slices.DeleteFunc(cfg.Pins, func(appid int) bool { return appid == c.AppID })
	if old > len(cfg.Pins) {
		return cfg.Save()
	}
	return nil
}

type Manager struct {
	AppID  int  `arg:"" placeholder:"APPID" help:"AppID to manage for updates."`
	Delete bool `short:"d" help:"Remove this AppID from managed apps instead."`
}

func (m *Manager) Run(cli *CLI) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if m.Delete {
		if _, ok := cfg.Games[m.AppID]; ok {
			delete(cfg.Games, m.AppID)
			return cfg.Save()
		}
	} else {
		if _, ok := cfg.Games[m.AppID]; !ok {
			name, err := getGameName(m.AppID)
			if err != nil {
				return fmt.Errorf("get game name for %d: %w", m.AppID, err)
			}
			cfg.Games[m.AppID] = name
			return cfg.Save()
		}
	}
	return nil
}

type CLI struct {
	ApiKey    string     `short:"a" env:"HUBCAP_KEY" help:"API key for authentication."`
	Verbose   bool       `short:"v" help:"Show detailed progress and manifest operations."`
	SLSconfig string     `short:"c" help:"Path to a custom SLSsteam config.yaml (for testing)."`
	Install   InstallCmd `cmd:"" aliases:"i" help:"Install an application."`
	Update    UpdateCmd  `cmd:"" aliases:"u" help:"Update installed applications."`
	Pin       PinCmd     `cmd:"" help:"Pin a game (temporarily ignored by update)"`
	Unpin     UnpinCmd   `cmd:"" help:"Unpin a game (updated again)."`
	Manage    Manager    `cmd:"" help:"Manage games managed by HuhCap (recognized and handled by updates)."`
}

func (c *CLI) Validate() error {
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if c.ApiKey == "" {
		c.ApiKey = cfg.ApiKey
	}
	if c.ApiKey == "" {
		return fmt.Errorf("Missing API key: pass --api (-a), set $HUBCAP_KEY, or run `huhcap --api <key>` to save one")
	}
	if cfg.ApiKey != c.ApiKey {
		cfg.ApiKey = c.ApiKey
		_ = cfg.Save()
	}
	return nil
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("huhcap"),
		kong.Description("CLI tool for managing applications."),
	)

	err := ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
