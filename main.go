package main

import (
	"fmt"
	"slices"

	"github.com/alecthomas/kong"
)

type InstallCmd struct {
	AppID int `arg:"" placeholder:"APPID" help:"AppId to install."`
}

func (c *InstallCmd) Run(cli *CLI) error {
	a, err := NewApp(cli.ApiKey, cli.Verbose, true, cli.SLSconfig)
	if err != nil {
		return err
	}
	name, err := getGameName(c.AppID)
	if err != nil {
		return err
	}
	err = a.InstallGame(c.AppID, name)
	if err == nil {
		if err := a.SaveSLSconfig(); err != nil {
			return err
		}
		if !slices.Contains(a.Config.Games, c.AppID) {
			a.Config.Games = append(a.Config.Games, c.AppID)
			return a.Config.Save()
		}
	}
	return err
}

type UpdateCmd struct {
	DryRun bool `help:"Just show what games are outdated."`
}

func (c *UpdateCmd) Run(cli *CLI) error {
	a, err := NewApp(cli.ApiKey, cli.Verbose, true, cli.SLSconfig)
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
	}
	return cfg.Save()
}

type UnpinCmd struct {
	AppID int `arg:"" placeholder:"APPID" help:"AppID to unpin."`
}

func (c *UnpinCmd) Run(cli *CLI) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	cfg.Pins = slices.DeleteFunc(cfg.Pins, func(appid int) bool { return appid == c.AppID })
	return cfg.Save()
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
		old := len(cfg.Games)
		cfg.Games = slices.DeleteFunc(cfg.Games, func(appid int) bool { return appid == m.AppID })
		if len(cfg.Games) < old {
			return cfg.Save()
		}
	} else {
		if !slices.Contains(cfg.Games, m.AppID) {
			cfg.Games = append(cfg.Games, m.AppID)
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
	Update    UpdateCmd  `cmd:"" help:"Update installed applications."`
	Pin       PinCmd     `cmd:"" help:"Pin a game (temporarily ignored by update)"`
	Unpin     UnpinCmd   `cmd:"" help:"Unpin a game (updated again)."`
	Manage    Manager    `cmd:"" help:"Manage games managed by HuhCap (recognized and handled by updates)."`
}

func (c *CLI) Validate() error {
	if c.ApiKey == "" {
		return fmt.Errorf("Missing API key: pass --api (-a) or set $HUBCAP_KEY")
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
