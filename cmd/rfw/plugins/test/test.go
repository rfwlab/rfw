package test

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
)

type plugin struct{}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "test" }

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) error {
	cfg := struct {
		Packages []string `json:"packages"`
	}{Packages: []string{"./..."}}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &cfg)
	}
	args := append([]string{"test"}, cfg.Packages...)
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	log.Printf("test: %s", strings.TrimSpace(string(output)))
	return err
}

func (p *plugin) ShouldRebuild(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}
