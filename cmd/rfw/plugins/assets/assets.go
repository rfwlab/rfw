package assets

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
	"github.com/rfwlab/rfw/v1/core"
)

type plugin struct {
	src string
	dst string
}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "assets" }

func (p *plugin) Install(a *core.App) {}

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) error {
	cfg := struct {
		Dir  string `json:"dir"`
		Dest string `json:"dest"`
	}{
		Dir:  "assets",
		Dest: "dist",
	}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &cfg)
	}
	p.src = cfg.Dir
	p.dst = cfg.Dest
	return filepath.Walk(cfg.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(cfg.Dir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(cfg.Dest, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
		log.Printf("assets: copied %s", target)
		return nil
	})
}

func (p *plugin) ShouldRebuild(path string) bool {
	return strings.HasPrefix(path, p.src)
}
