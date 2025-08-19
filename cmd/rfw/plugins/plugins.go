package plugins

import (
	"encoding/json"
	"fmt"

	"github.com/rfwlab/rfw/v1/core"
)

// Plugin defines the interface that build plugins must implement.
// It embeds the core.Plugin interface to allow plugins to also install runtime
// hooks when desired.
type Plugin interface {
	core.Plugin
	Name() string
	ShouldRebuild(path string) bool
}

var (
	registry = map[string]Plugin{}
	active   []Plugin
)

// Register adds a plugin to the registry.
func Register(p Plugin) { registry[p.Name()] = p }

// Install builds and activates a single plugin by name.
func Install(name string, raw json.RawMessage) error {
	if p, ok := registry[name]; ok {
		if err := p.Build(raw); err != nil {
			return fmt.Errorf("%s build failed: %w", p.Name(), err)
		}
		active = append(active, p)
		return nil
	}
	return fmt.Errorf("plugin %s not found", name)
}

// Configure installs plugins listed in the provided configuration map.
func Configure(cfg map[string]json.RawMessage) error {
	active = active[:0]
	for name, raw := range cfg {
		if err := Install(name, raw); err != nil {
			return err
		}
	}
	return nil
}

// NeedsRebuild reports whether any active plugin requires a rebuild for the given path.
func NeedsRebuild(path string) bool {
	for _, p := range active {
		if p.ShouldRebuild(path) {
			return true
		}
	}
	return false
}
