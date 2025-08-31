package plugins

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rfwlab/rfw/v1/core"
)

// Plugin defines the interface that build plugins must implement.
// It embeds the core.Plugin interface to allow plugins to also install runtime
// hooks when desired.
type Plugin interface {
	core.Plugin
	// Name returns a unique identifier for the plugin.
	Name() string
	// ShouldRebuild reports whether a change at the given path requires a rebuild.
	ShouldRebuild(path string) bool
	// Priority determines execution order; lower values run earlier.
	Priority() int
}

type entry struct {
	Plugin Plugin
	cfg    json.RawMessage
}

var (
	registry = map[string]Plugin{}
	active   []entry
)

// Register adds a plugin to the registry.
func Register(p Plugin) { registry[p.Name()] = p }

// Install builds and activates a single plugin by name.
func Install(name string, raw json.RawMessage) error {
	if p, ok := registry[name]; ok {
		active = append(active, entry{Plugin: p, cfg: raw})
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
	sort.SliceStable(active, func(i, j int) bool {
		return active[i].Plugin.Priority() < active[j].Plugin.Priority()
	})
	return nil
}

func PreBuild() error {
	for _, e := range active {
		if pb, ok := any(e.Plugin).(core.PreBuilder); ok {
			if err := pb.PreBuild(e.cfg); err != nil {
				return fmt.Errorf("%s prebuild failed: %w", e.Plugin.Name(), err)
			}
		}
	}
	return nil
}

func Build() error {
	for _, e := range active {
		if err := e.Plugin.Build(e.cfg); err != nil {
			return fmt.Errorf("%s build failed: %w", e.Plugin.Name(), err)
		}
	}
	return nil
}

func PostBuild() error {
	for _, e := range active {
		if pb, ok := any(e.Plugin).(core.PostBuilder); ok {
			if err := pb.PostBuild(e.cfg); err != nil {
				return fmt.Errorf("%s postbuild failed: %w", e.Plugin.Name(), err)
			}
		}
	}
	return nil
}

// NeedsRebuild reports whether any active plugin requires a rebuild for the given path.
func NeedsRebuild(path string) bool {
	for _, e := range active {
		if e.Plugin.ShouldRebuild(path) {
			return true
		}
	}
	return false
}
