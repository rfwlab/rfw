//go:build !js || !wasm

package core

// Component defines the minimal interface exposed to plugins in non-WASM builds.
type Component interface {
        Render() string
        GetName() string
        GetID() string
}

// ComponentRegistry holds constructors for components available to plugins.
var ComponentRegistry = map[string]func() Component{}

// RegisterComponent registers a component constructor for lookup by name.
func RegisterComponent(name string, constructor func() Component) {
        ComponentRegistry[name] = constructor
}

