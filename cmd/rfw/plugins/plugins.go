package plugins

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// Plugin defines the minimal interface that build plugins must implement.
// Lifecycle hooks like PreBuild, Build and PostBuild are detected
// automatically via reflection when present.
type Plugin interface {
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

// invoke executes the named lifecycle hook on the plugin if it exists.
func invoke(e entry, name string) error {
	m := reflect.ValueOf(e.Plugin).MethodByName(name)
	if !m.IsValid() {
		return nil
	}
	typ := m.Type()
	if typ.NumIn() != 1 || typ.In(0) != reflect.TypeOf(json.RawMessage{}) || typ.NumOut() != 1 || typ.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("%s has invalid signature for %s", e.Plugin.Name(), name)
	}
	out := m.Call([]reflect.Value{reflect.ValueOf(e.cfg)})
	if err, _ := out[0].Interface().(error); err != nil {
		return err
	}
	return nil
}

func PreBuild() error {
	for _, e := range active {
		if err := invoke(e, "PreBuild"); err != nil {
			return fmt.Errorf("%s prebuild failed: %w", e.Plugin.Name(), err)
		}
	}
	return nil
}

func Build() error {
	for _, e := range active {
		if err := invoke(e, "Build"); err != nil {
			return fmt.Errorf("%s build failed: %w", e.Plugin.Name(), err)
		}
	}
	return nil
}

func PostBuild() error {
	for _, e := range active {
		if err := invoke(e, "PostBuild"); err != nil {
			return fmt.Errorf("%s postbuild failed: %w", e.Plugin.Name(), err)
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
