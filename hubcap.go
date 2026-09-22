package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

// /api/v1/status/{app_id}
type StatusResponse struct {
	GameName    string `json:"game_name"`
	NeedsUpdate bool   `json:"needs_update"`
	Status      string `json:"status"`  // Check != "available"
	Message     string `json:"message"` // If there is a error
	Exists      bool   `json:"manifest_file_exists"`
}

// /api/v1/manifest/{app_id}/contents
type ContentResponse struct {
	ZipExists bool `json:"zip_exists"`
	Manifests []ManifestData
}
type ManifestData struct {
	Depot    string `json:"depot_id"`
	Manifest string `json:"manifest_id"`
}

// /api/v1/users/stats
type UsageResponse struct {
	Limit int `json:"daily_limit"`
	Usage int `json:"daily_usage"`
}

func (a *App) HubCapReq(url string, o ...any) *http.Request {
	if len(o) > 0 {
		url = fmt.Sprintf(url, o...)
	}
	req, _ := http.NewRequest("GET", "https://hubcapmanifest.com"+url, nil)
	req.Header.Set("Authorization", "Bearer "+a.ApiKey)
	return req
}

func (a *App) CheckUsage() (int, error) {
	var resp UsageResponse
	if err := HttpJson(a.HubCapReq("/api/v1/user/stats"), &resp); err != nil {
		return 0, err
	}
	n := resp.Limit - resp.Usage
	if n < 0 {
		return 0, errors.New("HubCap returned a negative usage limit")
	}
	return n, nil
}

func (a *App) CheckUpdate(appid int) (bool, []ManifestData, string, error) {
	var statusResp StatusResponse
	if err := HttpJson(a.HubCapReq("/api/v1/status/%d", appid), &statusResp); err != nil {
		return false, nil, "", err
	}
	if statusResp.Status != "available" || !statusResp.Exists {
		if statusResp.Message == "" {
			statusResp.Message = fmt.Sprintf("Game %d is not available in HubCap", appid)
		}
		return false, nil, "", errors.New(statusResp.Message)
	}
	// The manifests in HubCap's storage are out of date, so true
	if statusResp.NeedsUpdate {
		return true, nil, statusResp.GameName, nil
	}

	// Get the manifest gid's in HubCap's library to compare
	var contentResponse ContentResponse
	if err := HttpJson(a.HubCapReq("/api/v1/manifest/%d/contents", appid), &contentResponse); err != nil {
		return false, nil, "", err
	}
	if !contentResponse.ZipExists || len(contentResponse.Manifests) == 0 {
		return false, nil, "", fmt.Errorf("Invalid HubCap /content response")
	}
	return false, contentResponse.Manifests, statusResp.GameName, nil
}

func (a *App) InstallGame(appid int, name string) error {
	resp, err := HttpReq(a.HubCapReq("/api/v1/manifest/%d", appid))
	if err != nil {
		if resp == nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Game %s (%d) is not in HubCap", name, appid)
		}
		var js struct {
			Detail string `json:"detail"`
		}
		if jserr := json.NewDecoder(resp.Body).Decode(&js); jserr == nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, js.Detail)
		}
		return err
	}

	// Zip size is usually 0.5-60mb, so just read from memory
	buf, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to read response body: %w", err)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return fmt.Errorf("Failed to open zip: %w", err)
	}

	var luab []byte
	for _, file := range zipReader.File {
		switch filepath.Ext(file.Name) {
		case ".manifest":
			if err := a.copyManifest(file); err != nil {
				return err
			}
		case ".lua":
			if luab != nil {
				return fmt.Errorf("Zip has multiple .lua's (%s)", file.Name)
			}
			b, err := readZipFile(file)
			if err != nil {
				return fmt.Errorf("Failed reading .lua file %s: %w", file.Name, err)
			}
			luab = b
			a.Out.Verbose("Found Lua file")
		}
	}
	if luab == nil {
		return fmt.Errorf("No lua file found")
	}
	return a.parseLua(luab, appid, name)
}
