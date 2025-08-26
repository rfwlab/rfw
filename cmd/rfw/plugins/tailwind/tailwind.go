package tailwind

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
	"github.com/rfwlab/rfw/v1/core"
)

type plugin struct {
	output string
}

func init() {
	plugins.Register(&plugin{})
}

func (p *plugin) Name() string { return "tailwind" }

func (p *plugin) Install(a *core.App) {}

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) error {
	log.Printf("tailwind: starting build")
	bin, err := exec.LookPath("tailwindcss")
	if err != nil {
		log.Printf("tailwind: tailwindcss not found, please install it manually")
		return err
	}

	cfg := struct {
		Input  string   `json:"input"`
		Output string   `json:"output"`
		Minify bool     `json:"minify"`
		Args   []string `json:"args"`
	}{
		Input:  "index.css",
		Output: "tailwind.css",
		Minify: true,
	}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &cfg)
	}
	p.output = cfg.Output

	args := []string{"-i", cfg.Input, "-o", cfg.Output}
	if cfg.Minify {
		args = append(args, "--minify")
	}
	if len(cfg.Args) > 0 {
		args = append(args, cfg.Args...)
	}

	log.Printf("tailwind: running %s %s", bin, strings.Join(args, " "))
	cmd := exec.Command(bin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tailwind build failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	log.Printf("tailwind: build complete")
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	if strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, p.output) {
		log.Printf("tailwind: rebuild triggered by %s", path)
		return true
	}
	if strings.HasSuffix(path, ".rtml") || strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".go") {
		log.Printf("tailwind: rebuild triggered by %s", path)
		return true
	}
	return false
}
