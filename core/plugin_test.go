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

type depPlugin struct{ installed int }

func (p *depPlugin) Build(json.RawMessage) error { return nil }
func (p *depPlugin) Install(a *App)              { p.installed++ }
func (p *depPlugin) Name() string                { return "dep" }

type requiresPlugin struct{ dep *depPlugin }

func (p *requiresPlugin) Build(json.RawMessage) error { return nil }
func (p *requiresPlugin) Install(a *App)              {}
func (p *requiresPlugin) Name() string                { return "requires" }
func (p *requiresPlugin) Requires() []Plugin          { return []Plugin{p.dep} }

func TestRegisterPlugin_requires(t *testing.T) {
	app = newApp()
	dep := &depPlugin{}
	req := &requiresPlugin{dep: dep}
	RegisterPlugin(req)
	if dep.installed != 1 {
		t.Fatalf("expected dependency to install, got %d", dep.installed)
	}
	if !app.HasPlugin("dep") || !app.HasPlugin("requires") {
		t.Fatalf("expected both plugins to be registered")
	}
}

type optionalPlugin struct {
	dep    *depPlugin
	enable bool
}

func (p *optionalPlugin) Build(json.RawMessage) error { return nil }
func (p *optionalPlugin) Install(a *App)              {}
func (p *optionalPlugin) Name() string                { return "optional" }
func (p *optionalPlugin) Optional() []Plugin {
	if !p.enable {
		return nil
	}
	return []Plugin{p.dep}
}

func TestRegisterPlugin_optional(t *testing.T) {
	app = newApp()
	dep := &depPlugin{}
	opt := &optionalPlugin{dep: dep, enable: true}
	RegisterPlugin(opt)
	if dep.installed != 1 {
		t.Fatalf("expected optional dependency to install")
	}

	app = newApp()
	dep2 := &depPlugin{}
	opt2 := &optionalPlugin{dep: dep2, enable: false}
	RegisterPlugin(opt2)
	if dep2.installed != 0 {
		t.Fatalf("expected disabled optional dependency not to install")
	}
}
