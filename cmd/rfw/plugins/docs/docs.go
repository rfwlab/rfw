//go:build !js

// Package docs registers documentation asset processing.
package docs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/v2/cmd/rfw/plugins"
)

type plugin struct {
	src string
}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "docs" }

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) (err error) {
	cfg := struct {
		Dir  string `json:"dir"`
		Dest string `json:"dest"`
	}{
		Dir:  "articles",
		Dest: filepath.Join("build", "static"),
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return fmt.Errorf("decode docs plugin config: %w", err)
		}
	}
	sourcePath, err := docsProjectDirectory(cfg.Dir)
	if err != nil {
		return fmt.Errorf("invalid docs source: %w", err)
	}
	destinationPath, err := docsProjectDirectory(cfg.Dest)
	if err != nil {
		return fmt.Errorf("invalid docs destination: %w", err)
	}
	p.src = sourcePath
	base := filepath.Base(sourcePath)
	destRoot := filepath.Join(destinationPath, base)

	projectRoot, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := projectRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := projectRoot.MkdirAll(destRoot, 0o755); err != nil {
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
	targetRoot, err := projectRoot.OpenRoot(destRoot)
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
		closeErr := errors.Join(out.Close(), in.Close())
		return errors.Join(copyErr, closeErr)
	})
	return err
}

func (p *plugin) ShouldRebuild(path string) bool {
	relative, err := filepath.Rel(p.src, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func docsProjectDirectory(path string) (string, error) {
	if path == "" || filepath.IsAbs(path) {
		return "", fmt.Errorf("path %q must be project-relative", path)
	}
	cleaned := filepath.Clean(path)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes the project", path)
	}
	return cleaned, nil
}
