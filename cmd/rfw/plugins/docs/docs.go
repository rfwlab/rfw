package docs

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
)

type plugin struct {
	src string
}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "docs" }

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) error {
	cfg := struct {
		Dir  string `json:"dir"`
		Dest string `json:"dest"`
	}{
		Dir:  "articles",
		Dest: filepath.Join("build", "static"),
	}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &cfg)
	}
	p.src = cfg.Dir
	base := filepath.Base(cfg.Dir)
	destRoot := filepath.Join(cfg.Dest, base)
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
		target := filepath.Join(destRoot, rel)
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
		return out.Close()
	})
}

func (p *plugin) ShouldRebuild(path string) bool {
	return strings.HasPrefix(path, p.src)
}
