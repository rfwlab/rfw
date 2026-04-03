package plugins

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Plugin defines the minimal interface that build plugins must implement.
type Plugin interface {
	Name() string
	ShouldRebuild(path string) bool
	Priority() int
}

// PreBuilder is implemented by plugins that run before the build.
type PreBuilder interface{ PreBuild(json.RawMessage) error }

// Builder is implemented by plugins that participate in the build step.
type Builder interface{ Build(json.RawMessage) error }

// PostBuilder is implemented by plugins that run after the build completes.
type PostBuilder interface{ PostBuild(json.RawMessage) error }

type entry struct {
	Plugin Plugin
	cfg    json.RawMessage
}

var (
	registry = map[string]Plugin{}
	active   []entry
)

// Info reports the name and configuration of an active plugin.
type Info struct {
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

// Active returns the list of configured plugins in execution order.
func Active() []Info {
	out := make([]Info, len(active))
	for i, e := range active {
		out[i] = Info{Name: e.Plugin.Name(), Config: e.cfg}
	}
	return out
}

// Register adds a plugin to the registry.
func Register(p Plugin) { registry[p.Name()] = p }

// Install activates a single plugin by name.
func Install(name string, raw json.RawMessage) error {
	if p, ok := registry[name]; ok {
		active = append(active, entry{Plugin: p, cfg: raw})
		return nil
	}
	return fmt.Errorf("plugin %s not found", name)
}

// Configure installs plugins from the provided configuration map.
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

// NeedsRebuild reports whether any active plugin requires a rebuild for the given path.
func NeedsRebuild(path string) bool {
	for _, e := range active {
		if e.Plugin.ShouldRebuild(path) {
			return true
		}
	}
	return false
}

// PreBuild calls PreBuild on all plugins that implement PreBuilder.
func PreBuild() error {
	for _, e := range active {
		if pb, ok := e.Plugin.(PreBuilder); ok {
			if err := pb.PreBuild(e.cfg); err != nil {
				return fmt.Errorf("%s prebuild: %w", e.Plugin.Name(), err)
			}
		}
	}
	return nil
}

// Build calls Build on all plugins that implement Builder.
func Build() error {
	for _, e := range active {
		if b, ok := e.Plugin.(Builder); ok {
			if err := b.Build(e.cfg); err != nil {
				return fmt.Errorf("%s build: %w", e.Plugin.Name(), err)
			}
		}
	}
	return nil
}

// PostBuild calls PostBuild on all plugins that implement PostBuilder.
func PostBuild() error {
	for _, e := range active {
		if pb, ok := e.Plugin.(PostBuilder); ok {
			if err := pb.PostBuild(e.cfg); err != nil {
				return fmt.Errorf("%s postbuild: %w", e.Plugin.Name(), err)
			}
		}
	}
	return nil
}
