package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var Client = &http.Client{Timeout: 20 * time.Second}

func HttpReq(req *http.Request) (*http.Response, error) {
	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
	return resp, nil
}

func HttpJson(req *http.Request, v any) error {
	resp, err := HttpReq(req)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("Malformed response: %w", err)
	}
	return nil
}

func getGameName(appid int) (string, error) {
	resp, err := Client.Get(fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d", appid))
	if err != nil {
		return "", fmt.Errorf("Request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
	var s map[string]struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return "", fmt.Errorf("Malformed response: %w", err)
	}
	n := s[fmt.Sprintf("%d", appid)].Data.Name
	if n == "" {
		return "", errors.New("Name is empty, malformed response")
	}
	return n, nil
}

func (a *App) copyManifest(file *zip.File) error {
	n := strings.TrimSuffix(filepath.Base(file.Name), ".manifest")
	zf, err := file.Open()
	if err != nil {
		return fmt.Errorf("Failed to open manifest: %w", err)
	}
	defer zf.Close()
	dst := filepath.Join(a.Depotcache, file.Name)
	df, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if errors.Is(err, os.ErrExist) {
		a.Out.Verbose("↷ Skipping existing manifest %s", n)
		return nil
	} else if err != nil {
		return fmt.Errorf("Failed to open manifest: %w", err)
	}
	defer df.Close()
	_, err = io.Copy(df, zf)
	if err != nil {
		return fmt.Errorf("Failed to copy manifest: %w", err)
	}
	a.Out.Verbose("+ Copied manifest %s", n)
	return nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	zf, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer zf.Close()
	return io.ReadAll(zf)
}
