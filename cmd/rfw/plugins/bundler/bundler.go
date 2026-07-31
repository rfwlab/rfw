//go:build !js

// Package bundler registers post-build asset minification.
package bundler

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/v2/cmd/rfw/logging"
	"github.com/rfwlab/rfw/v2/cmd/rfw/plugins"
	"github.com/rfwlab/rfw/v2/cmd/rfw/utils"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

type plugin struct{}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "bundler" }

func (p *plugin) Priority() int { return 10 }

func (p *plugin) PostBuild(_ json.RawMessage) (err error) {
	if utils.IsDebug() {
		logging.Log.Info("skipped in debug mode", logging.F("plugin", "bundler"))
		return nil
	}

	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)

	buildDir := "build"
	root, err := os.OpenRoot(buildDir)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := root.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := filepath.Walk(buildDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(buildDir, path)
		if err != nil {
			return err
		}
		ext := filepath.Ext(path)
		var media string
		var data []byte
		switch ext {
		case ".js":
			media = "text/javascript"
		case ".css":
			data, err = readRootFile(root, rel)
			if err != nil {
				return err
			}
			if isTailwindCSSData(data) {
				logging.Log.Info("skipping tailwind css", logging.F("plugin", "bundler"), logging.F("path", path))
				return nil
			}
			media = "text/css"
		case ".html":
			media = "text/html"
		default:
			return nil
		}
		if data == nil {
			data, err = readRootFile(root, rel)
			if err != nil {
				return err
			}
		}
		out, err := m.Bytes(media, data)
		if err != nil {
			return err
		}
		file, err := root.OpenFile(rel, os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(out)
		return errors.Join(writeErr, file.Close())
	}); err != nil {
		return err
	}
	logging.Log.Info("build complete", logging.F("plugin", "bundler"))
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, "build/") {
		return false
	}
	return strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".html")
}

func readRootFile(root *os.Root, path string) ([]byte, error) {
	file, err := root.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(file)
	return data, errors.Join(readErr, file.Close())
}

func isTailwindCSSData(data []byte) bool {
	src := string(data)
	return strings.Contains(src, "@tailwind") || strings.Contains(src, "tailwindcss")
}
