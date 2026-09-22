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
	if err = a.InstallGame(c.AppID, ""); err == nil {
		if err := a.SaveSLSconfig(); err != nil {
			return err
		}
		if !slices.Contains(a.Config.Games, c.AppID) {
			a.Config.Games = append(a.Config.Games, c.AppID)
			return a.Config.Save()
		}
	}
	return nil
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
	AppID int `arg:"" placeholder:"APPID" help:"Application ID to pin."`
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
	AppID int `arg:"" placeholder:"APPID" help:"Application ID to unpin."`
}

func (c *UnpinCmd) Run(cli *CLI) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	cfg.Pins = slices.DeleteFunc(cfg.Pins, func(appid int) bool { return appid == c.AppID })
	return cfg.Save()
}

type CLI struct {
	ApiKey    string     `short:"a" env:"HUBCAP_KEY" help:"API key for authentication."`
	Verbose   bool       `short:"v" help:"Show detailed progress and manifest operations."`
	SLSconfig string     `short:"c" help:"Path to a custom SLSsteam config.yaml (for testing)."`
	Install   InstallCmd `cmd:"" aliases:"i" help:"Install an application."`
	Update    UpdateCmd  `cmd:"" help:"Update installed applications."`
	Pin       PinCmd     `cmd:"" help:"Pin a game (ignore by update)"`
	Unpin     UnpinCmd   `cmd:"" help:"Unpin a game (updated again)."`
}

func (c *CLI) Validate() error {
	if c.ApiKey == "" {
		return fmt.Errorf("missing API key: pass --api (-a) or set $HUBCAP_KEY")
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
