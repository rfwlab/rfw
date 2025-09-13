//go:build js && wasm

package core

import (
	"encoding/json"
	"testing"
)

type namedTestPlugin struct{ installed int }

func (p *namedTestPlugin) Build(json.RawMessage) error { return nil }
func (p *namedTestPlugin) Install(a *App)              { p.installed++ }
func (p *namedTestPlugin) Name() string                { return "named-test" }

func TestRegisterPlugin_dedup(t *testing.T) {
	app = newApp()
	p1 := &namedTestPlugin{}
	RegisterPlugin(p1)
	if p1.installed != 1 {
		t.Fatalf("expected first plugin to install once, got %d", p1.installed)
	}
	p2 := &namedTestPlugin{}
	RegisterPlugin(p2)
	if p2.installed != 0 {
		t.Fatalf("expected second plugin not to install, got %d", p2.installed)
	}
	if !app.HasPlugin("named-test") {
		t.Fatalf("expected HasPlugin to return true")
	}
}
