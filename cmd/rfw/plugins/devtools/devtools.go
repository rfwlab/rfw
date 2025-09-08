package devtools

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
)

//go:embed devtools.js
var devtoolsJS []byte

type plugin struct{}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "devtools" }

func (p *plugin) Priority() int { return 100 }

func (p *plugin) PreBuild(json.RawMessage) error {
	stub := []byte("package main\nimport _ \"github.com/rfwlab/rfw/v1/devtools\"\n")
	return os.WriteFile("rfw_devtools.go", stub, 0o644)
}

func (p *plugin) PostBuild(json.RawMessage) error {
	dst := filepath.Join("build", "client", "rfw-devtools.js")
	if err := os.WriteFile(dst, devtoolsJS, 0o644); err != nil {
		return err
	}
	index := filepath.Join("build", "client", "index.html")
	data, err := os.ReadFile(index)
	if err != nil {
		return err
	}
	if !bytes.Contains(data, []byte("rfw-devtools.js")) {
		injection := []byte("\n<script type=\"module\" src=\"/rfw-devtools.js\"></script>\n")
		data = bytes.Replace(data, []byte("</body>"), append(injection, []byte("</body>")...), 1)
		if err := os.WriteFile(index, data, 0o644); err != nil {
			return err
		}
	}
	_ = os.Remove("rfw_devtools.go")
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	return strings.HasSuffix(path, "devtools.js")
}
