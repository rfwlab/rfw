package plugins

import (
	"encoding/json"
	"fmt"
)

// Plugin defines the interface that build plugins must implement.
type Plugin interface {
	Name() string
	Build(cfg json.RawMessage) error
	ShouldRebuild(path string) bool
}

var (
	registry = map[string]Plugin{}
	active   []Plugin
)

// Register adds a plugin to the registry.
func Register(p Plugin) {
	registry[p.Name()] = p
}

// BuildFromConfig executes the Build method for plugins listed in the manifest.
func BuildFromConfig(cfg map[string]json.RawMessage) error {
	active = active[:0]
	for name, raw := range cfg {
		if p, ok := registry[name]; ok {
			if err := p.Build(raw); err != nil {
				return fmt.Errorf("%s build failed: %w", p.Name(), err)
			}
			active = append(active, p)
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
