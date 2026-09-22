package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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
