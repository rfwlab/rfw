//go:build js && wasm

package core

// DevMode enables additional runtime checks and warnings for development.
// It is disabled by default.
var DevMode bool

// SetDevMode toggles development mode features.
func SetDevMode(enabled bool) { DevMode = enabled }

type Component interface {
	Render() string
	Mount()
	Unmount()
	OnMount()
	OnUnmount()
	GetName() string
	GetID() string
	SetSlots(map[string]string)
}

// ComponentRegistry holds constructors for components that can be loaded on-demand.
// Components can register themselves with a unique name and a constructor function
// returning a new instance of the component. This allows templates to reference
// components by name using runtime selection attributes like `rt-is`.
var ComponentRegistry = map[string]func() Component{}

// RegisterComponent registers a component constructor under the provided name.
// When a template references the name via `rt-is`, the constructor will be
// invoked to create a new component instance at render time.
func RegisterComponent(name string, constructor func() Component) {
	ComponentRegistry[name] = constructor
}

// LoadComponent retrieves a component by name using the registry. If no
// component is registered under that name, nil is returned.
func LoadComponent(name string) Component {
	if ctor, ok := ComponentRegistry[name]; ok {
		return ctor()
	}
	return nil
}
