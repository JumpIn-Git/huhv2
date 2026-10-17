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
	"regexp"
	"strconv"
	"strings"
	"time"
)

var steamURLRegex = regexp.MustCompile(`^(?:https?://)?store\.steampowered\.com/app/(\d+)(?:/([^/]+))?/?$`)

func (a *App) parseAppID(g string) (int, string, error) {
	if n, err := strconv.Atoi(g); err == nil {
		return n, "", nil
	}
	if matches := steamURLRegex.FindStringSubmatch(g); len(matches) == 3 {
		// [{g}, int, string (can be "")]
		if n, err := strconv.Atoi(matches[1]); err == nil {
			return n, matches[2], nil
		}
	}
	return a.searchGameName(g)
}

var Client = &http.Client{Timeout: 20 * time.Second}

func HttpReq(req *http.Request) (*http.Response, error) {
	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", req.URL, err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, req.URL)
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
	url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d", appid)
	resp, err := Client.Get(url)
	if err != nil {
		return "", fmt.Errorf("Steam request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Steam returned status %d for %d", resp.StatusCode, appid)
	}
	var s map[string]struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return "", fmt.Errorf("malformed Steam response for %d: %w", appid, err)
	}
	n := s[fmt.Sprintf("%d", appid)].Data.Name
	if n == "" {
		return "", fmt.Errorf("Steam returned empty name for %d", appid)
	}
	return n, nil
}

func (a *App) copyManifest(file *zip.File) error {
	n := strings.TrimSuffix(filepath.Base(file.Name), ".manifest")
	zf, err := file.Open()
	if err != nil {
		return fmt.Errorf("open manifest %s: %w", file.Name, err)
	}
	defer zf.Close()
	dst := filepath.Join(a.Depotcache, file.Name)
	df, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if errors.Is(err, os.ErrExist) {
		a.Out.Verbose("↷ Skipping existing manifest %s", n)
		return nil
	} else if err != nil {
		return fmt.Errorf("create manifest %s: %w", dst, err)
	}
	defer df.Close()
	if _, err = io.Copy(df, zf); err != nil {
		return fmt.Errorf("copy manifest %s: %w", n, err)
	}
	a.Out.Verbose("+ Copied manifest %s", n)
	return nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	zf, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open zip entry %s: %w", file.Name, err)
	}
	defer zf.Close()
	b, err := io.ReadAll(zf)
	if err != nil {
		return nil, fmt.Errorf("read zip entry %s: %w", file.Name, err)
	}
	return b, nil
}
