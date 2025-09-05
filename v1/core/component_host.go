//go:build !js || !wasm

package core

import "fmt"

// Component defines the minimal interface exposed to plugins in non-WASM builds.
type Component interface {
	Render() string
	GetName() string
	GetID() string
}

// ComponentRegistry holds constructors for components available to plugins.
var ComponentRegistry = map[string]func() Component{}

// RegisterComponent registers a component constructor for lookup by name. It
// returns an error if a component with the same name has already been
// registered and logs a warning.
func RegisterComponent(name string, constructor func() Component) error {
	if _, exists := ComponentRegistry[name]; exists {
		Log().Warn("component %s already registered", name)
		return fmt.Errorf("component %s already registered", name)
	}
	ComponentRegistry[name] = constructor
	return nil
}
