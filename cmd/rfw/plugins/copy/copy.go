//go:build !js

// Package copy registers configurable build file copies.
package copy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rfwlab/rfw/v2/cmd/rfw/logging"
	"github.com/rfwlab/rfw/v2/cmd/rfw/plugins"
)

type rule struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type plugin struct {
	rules []rule
}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "copy" }

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) (err error) {
	cfg := struct {
		Files []rule `json:"files"`
	}{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return err
		}
	}
	p.rules = cfg.Files
	projectRoot, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := projectRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	for _, r := range p.rules {
		if err := validateRule(r); err != nil {
			return err
		}
		matches, err := doublestar.Glob(os.DirFS("."), r.From)
		if err != nil {
			return err
		}
		base, _ := doublestar.SplitPattern(r.From)
		base = filepath.FromSlash(base)
		for _, m := range matches {
			path := filepath.FromSlash(m)
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				continue
			}
			rel, err := filepath.Rel(base, path)
			if err != nil {
				return err
			}
			dst := filepath.Join(r.To, rel)
			destinationDirectory := filepath.Dir(dst)
			if err := projectRoot.MkdirAll(destinationDirectory, 0o755); err != nil {
				return err
			}
			if err := copyFile(projectRoot, path, dst); err != nil {
				return err
			}
			logging.Log.Info("copied file", logging.F("plugin", "copy"), logging.F("path", dst))
		}
	}
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	for _, r := range p.rules {
		if ok, _ := doublestar.PathMatch(r.From, path); ok {
			return true
		}
	}
	return false
}

func validateRule(copyRule rule) error {
	for field, value := range map[string]string{"from": copyRule.From, "to": copyRule.To} {
		if value == "" || filepath.IsAbs(value) {
			return fmt.Errorf("copy %s path %q must be project-relative", field, value)
		}
		for _, segment := range strings.FieldsFunc(filepath.Clean(value), func(char rune) bool {
			return char == '/' || char == '\\'
		}) {
			if segment == ".." {
				return fmt.Errorf("copy %s path %q escapes the project", field, value)
			}
		}
	}
	return nil
}

func copyFile(projectRoot *os.Root, src, dst string) error {
	in, err := projectRoot.Open(src)
	if err != nil {
		return err
	}
	if err := projectRoot.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.Join(err, in.Close())
	}
	out, err := projectRoot.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return errors.Join(err, in.Close())
	}
	_, copyErr := io.Copy(out, in)
	closeErr := errors.Join(out.Close(), in.Close())
	return errors.Join(copyErr, closeErr)
}
