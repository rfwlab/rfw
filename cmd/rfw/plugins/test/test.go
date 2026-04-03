package test

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/logging"
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
	out := strings.TrimSpace(string(output))
	if err != nil {
		logging.Log.Error("go test failed", logging.F("plugin", "test"), logging.F("output", out), logging.F("error", err.Error()))
		return err
	}
	logging.Log.Info("go test ok", logging.F("plugin", "test"), logging.F("output", out))
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}
