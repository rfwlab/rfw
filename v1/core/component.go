//go:build js && wasm

package core

import (
	"fmt"
	"sync"
)

// DevMode enables additional runtime checks and warnings for development.
// It is disabled by default.
var DevMode bool

// SetDevMode toggles development mode features.
func SetDevMode(enabled bool) {
        DevMode = enabled
        if enabled {
                startDevTemplateWatcher()
        }
}

type Component interface {
	Render() string
	Mount()
	Unmount()
	OnMount()
	OnUnmount()
	GetName() string
	GetID() string
	SetSlots(map[string]any)
}

// ComponentRegistry holds constructors for components that can be loaded on-demand.
// Components can register themselves with a unique name and a constructor function
// returning a new instance of the component. This allows templates to reference
// components by name using runtime selection attributes like `rt-is`.
var (
	ComponentRegistry   = map[string]func() Component{}
	componentRegistryMu sync.RWMutex
)

// RegisterComponent registers a component constructor under the provided name.
// When a template references the name via `rt-is`, the constructor will be
// invoked to create a new component instance at render time. It returns an
// error if a component with the same name is already registered and logs a
// warning.
func RegisterComponent(name string, constructor func() Component) error {
	componentRegistryMu.Lock()
	defer componentRegistryMu.Unlock()
	if _, exists := ComponentRegistry[name]; exists {
		Log().Warn("component %s already registered", name)
		return fmt.Errorf("component %s already registered", name)
	}
	ComponentRegistry[name] = constructor
	return nil
}

// LoadComponent retrieves a component by name using the registry. If no
// component is registered under that name, nil is returned.
func LoadComponent(name string) Component {
	componentRegistryMu.RLock()
	ctor, ok := ComponentRegistry[name]
	componentRegistryMu.RUnlock()
	if ok {
		return ctor()
	}
	return nil
}

// NewComponent creates an HTMLComponent initialized with the provided
// template and props. It sets itself as the underlying component and
// performs initialization with the default store.
func NewComponent(name string, templateFS []byte, props map[string]any) *HTMLComponent {
	c := NewHTMLComponent(name, templateFS, props)
	c.SetComponent(c)
	c.Init(nil)
	return c
}

// NewComponentWith creates an HTMLComponent and binds it to the given
// component implementation. This is useful when embedding HTMLComponent
// inside another struct to override lifecycle hooks.
func NewComponentWith[T Component](name string, templateFS []byte, props map[string]any, self T) *HTMLComponent {
	c := NewHTMLComponent(name, templateFS, props)
	if any(self) != nil {
		c.SetComponent(self)
	} else {
		c.SetComponent(c)
	}
	c.Init(nil)
	return c
}
