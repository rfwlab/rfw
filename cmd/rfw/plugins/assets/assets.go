//go:build !js

// Package assets registers the asset-copy build plugin.
package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/v2/cmd/rfw/logging"
	"github.com/rfwlab/rfw/v2/cmd/rfw/plugins"
)

type plugin struct {
	src string
	dst string
}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "assets" }

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) (err error) {
	cfg := struct {
		Dir  string `json:"dir"`
		Dest string `json:"dest"`
	}{
		Dir:  "assets",
		Dest: "dist",
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return fmt.Errorf("decode assets plugin config: %w", err)
		}
	}
	sourcePath, err := projectDirectory(cfg.Dir)
	if err != nil {
		return fmt.Errorf("invalid assets source: %w", err)
	}
	destinationPath, err := projectDirectory(cfg.Dest)
	if err != nil {
		return fmt.Errorf("invalid assets destination: %w", err)
	}
	p.src = sourcePath
	p.dst = destinationPath

	projectRoot, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := projectRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := projectRoot.MkdirAll(destinationPath, 0o755); err != nil {
		return err
	}
	sourceRoot, err := projectRoot.OpenRoot(sourcePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sourceRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	targetRoot, err := projectRoot.OpenRoot(destinationPath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := targetRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	err = fs.WalkDir(sourceRoot.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == "." {
				return nil
			}
			if err := targetRoot.MkdirAll(path, 0o755); err != nil {
				return err
			}
			return nil
		}
		in, err := sourceRoot.Open(path)
		if err != nil {
			return err
		}
		if err := targetRoot.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return errors.Join(err, in.Close())
		}
		out, err := targetRoot.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			return errors.Join(err, in.Close())
		}
		_, copyErr := io.Copy(out, in)
		outCloseErr := out.Close()
		inCloseErr := in.Close()
		if err := errors.Join(copyErr, outCloseErr, inCloseErr); err != nil {
			return err
		}
		logging.Log.Info("copied file", logging.F("plugin", "assets"), logging.F("path", filepath.Join(destinationPath, path)))
		return nil
	})
	return err
}

func (p *plugin) ShouldRebuild(path string) bool {
	relative, err := filepath.Rel(p.src, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func projectDirectory(path string) (string, error) {
	if path == "" || filepath.IsAbs(path) {
		return "", fmt.Errorf("path %q must be project-relative", path)
	}
	cleaned := filepath.Clean(path)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes the project", path)
	}
	return cleaned, nil
}
