package main

import (
	"fmt"
	"slices"

	"charm.land/huh/v2"
)

type foundGame struct {
	AppID int
	Name  string
}

func (fg foundGame) String() string {
	return fmt.Sprintf("%s (%d)", fg.Name, fg.AppID)
}

func (a *App) Update(dryrun bool) error {
	found := make([]foundGame, 0)
	manifestMap := make(map[string]string, len(a.ManifestIds.Content)/2)
	for i := 0; i < len(a.ManifestIds.Content); i += 2 {
		manifestMap[a.ManifestIds.Content[i].Value] = a.ManifestIds.Content[i+1].Value
	}

	for i, appid := range a.Config.Games {
		if slices.Contains(a.Config.Pins, appid) {
			a.Out.Info("Skipping pinned %d (%d/%d)", appid, i+1, len(a.Config.Games))
			continue
		}
		a.Out.Info("Checking %d/%d", i+1, len(a.Config.Games))
		upd, data, name, err := a.CheckUpdate(appid)
		if err != nil {
			return err
		} else if !upd {
			for _, d := range data {
				if val, exists := manifestMap[d.Depot]; !exists || val != d.Manifest {
					upd = true
					break
				}
			}
		}
		if upd {
			found = append(found, foundGame{appid, name})
		}
	}
	if len(found) == 0 {
		a.Out.Info("No updates!")
		return nil
	} else if dryrun {
		a.Out.Info("Found these updates: %v", found)
		return nil
	}

	limit, err := a.CheckUsage()
	if err != nil {
		return err
	}
	var selected []foundGame
	s := huh.NewMultiSelect[foundGame]().
		Title("Which games do you want to update?").
		Options(huh.NewOptions(found...)...).
		Value(&selected)
	if limit >= len(found) {
		s.Description("Space to toggle.")
	} else {
		s.Limit(limit).Description(fmt.Sprintf("Space to toggle, you can update %d.", limit))
	}
	if err := s.Run(); err != nil {
		return err
	}

	for i, g := range selected {
		if err := a.InstallGame(g.AppID, g.Name); err != nil {
			if i > 0 {
				if saveErr := a.SaveSLSconfig(); saveErr != nil {
					return fmt.Errorf("%w (%w)", err, saveErr)
				}
			}
			return err
		}
	}
	return a.SaveSLSconfig()
}
