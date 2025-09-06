package plugins

import (
	"encoding/json"
	"testing"
)

// mockPlugin is a simple implementation of the Plugin interface used for
// testing the plugin registry and lifecycle.
type mockPlugin struct {
	name     string
	priority int
	rebuild  string
	pre      bool
	build    bool
	post     bool
}

func (m *mockPlugin) Name() string                { return m.name }
func (m *mockPlugin) Priority() int               { return m.priority }
func (m *mockPlugin) ShouldRebuild(p string) bool { return p == m.rebuild }
func (m *mockPlugin) PreBuild(json.RawMessage) error {
	m.pre = true
	return nil
}
func (m *mockPlugin) Build(json.RawMessage) error {
	m.build = true
	return nil
}
func (m *mockPlugin) PostBuild(json.RawMessage) error {
	m.post = true
	return nil
}

// TestLifecycle verifies registration, ordering and lifecycle invocation of
// plugins as well as the NeedsRebuild helper.
func TestLifecycle(t *testing.T) {
	// Reset global state.
	registry = map[string]Plugin{}
	active = nil

	p1 := &mockPlugin{name: "p1", priority: 1, rebuild: "a.go"}
	p0 := &mockPlugin{name: "p0", priority: 0}
	Register(p1)
	Register(p0)

	cfg := map[string]json.RawMessage{"p1": nil, "p0": nil}
	if err := Configure(cfg); err != nil {
		t.Fatalf("configure: %v", err)
	}
	if len(active) != 2 || active[0].Plugin != p0 || active[1].Plugin != p1 {
		t.Fatalf("plugins not sorted by priority: %#v", active)
	}

	if err := PreBuild(); err != nil {
		t.Fatalf("prebuild: %v", err)
	}
	if !p1.pre || !p0.pre {
		t.Fatalf("prebuild not invoked")
	}
	if err := Build(); err != nil {
		t.Fatalf("build: %v", err)
	}
	if !p1.build || !p0.build {
		t.Fatalf("build not invoked")
	}
	if err := PostBuild(); err != nil {
		t.Fatalf("postbuild: %v", err)
	}
	if !p1.post || !p0.post {
		t.Fatalf("postbuild not invoked")
	}

	if !NeedsRebuild("a.go") {
		t.Fatalf("expected rebuild for a.go")
	}
	if NeedsRebuild("b.go") {
		t.Fatalf("unexpected rebuild for b.go")
	}
}
