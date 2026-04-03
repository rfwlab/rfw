package tailwind

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/logging"
	"github.com/rfwlab/rfw/cmd/rfw/plugins"
)

type plugin struct {
	output string
}

func init() {
	plugins.Register(&plugin{})
}

func (p *plugin) Name() string { return "tailwind" }

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) error {
	logging.Log.Info("starting build", logging.F("plugin", "tailwind"))
	bin, err := exec.LookPath("tailwindcss")
	if err != nil {
		logging.Log.Warn("tailwindcss not found, please install it manually", logging.F("plugin", "tailwind"))
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

	logging.Log.Info("running command", logging.F("plugin", "tailwind"), logging.F("bin", bin), logging.F("args", strings.Join(args, " ")))
	cmd := exec.Command(bin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tailwind build failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	logging.Log.Info("build complete", logging.F("plugin", "tailwind"))
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	if strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, p.output) {
		logging.Log.Info("rebuild triggered", logging.F("plugin", "tailwind"), logging.F("path", path))
		return true
	}
	if strings.HasSuffix(path, ".rtml") || strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".go") {
		logging.Log.Info("rebuild triggered", logging.F("plugin", "tailwind"), logging.F("path", path))
		return true
	}
	return false
}
