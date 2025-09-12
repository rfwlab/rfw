package copy

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rfwlab/rfw/cmd/rfw/plugins"
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

func (p *plugin) Build(raw json.RawMessage) error {
	cfg := struct {
		Files []rule `json:"files"`
	}{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return err
		}
	}
	p.rules = cfg.Files
	for _, r := range p.rules {
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
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			if err := copyFile(path, dst); err != nil {
				return err
			}
			log.Printf("copy: copied %s", dst)
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
