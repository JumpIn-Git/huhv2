package main

import (
	"fmt"
	"slices"

	"github.com/alecthomas/kong"
)

type InstallCmd struct {
	AppIDs []string `arg:"" required:"" placeholder:"APPID" help:"Games to install, can be a appid, store page or game name to search."`
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
			return err
		}
		if name == "" {
			name, err = getGameName(appid)
			if err != nil {
				return err
			}
		}
		found = append(found, foundGame{appid, name})
	}

	// TODO add limit check
	var upd bool
	for _, g := range found {
		if err := a.InstallGame(g.AppID, g.Name); err != nil {
			if upd {
				if saveErr := a.Config.Save(); saveErr != nil {
					err = fmt.Errorf("%w (%w)", err, saveErr)
				}
			}
			return err
		}
		if !slices.Contains(a.Config.Games, g.AppID) {
			a.Config.Games = append(a.Config.Games, g.AppID)
			upd = true
		}
		a.Out.Info("Installed %s!", g.Name)
	}
	if upd {
		return a.Config.Save()
	}
	return nil
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
		old := len(cfg.Games)
		cfg.Games = slices.DeleteFunc(cfg.Pins, func(appid int) bool { return appid == m.AppID })
		if old > len(cfg.Games) {
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

// TODO add api key manager, if both config api key and argument is emtpy error
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
