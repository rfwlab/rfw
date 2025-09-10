package bundler

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
	"github.com/rfwlab/rfw/cmd/rfw/utils"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

type plugin struct{}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "bundler" }

func (p *plugin) Priority() int { return 10 }

func (p *plugin) PostBuild(raw json.RawMessage) error {
	if utils.IsDebug() {
		log.Printf("bundler: skipped in debug mode")
		return nil
	}

	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)

	buildDir := "build"
	if err := filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		var media string
		switch ext {
		case ".js":
			media = "text/javascript"
		case ".css":
			if isTailwindCSS(path) {
				log.Printf("bundler: skipping tailwind css %s", path)
				return nil
			}
			media = "text/css"
		case ".html":
			media = "text/html"
		default:
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out, err := m.Bytes(media, data)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, out, 0o644); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	log.Printf("bundler: build complete")
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, "build/") {
		return false
	}
	return strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".html")
}

func isTailwindCSS(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	src := string(data)
	return strings.Contains(src, "@tailwind") || strings.Contains(src, "tailwindcss")
}
