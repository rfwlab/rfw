package plugins

import "fmt"

// Plugin defines the interface that build plugins must implement.
type Plugin interface {
	Name() string
	Build() error
	ShouldRebuild(path string) bool
}

var registered []Plugin

// Register adds a plugin to the registry.
func Register(p Plugin) {
	registered = append(registered, p)
}

// BuildAll executes the Build method for all registered plugins.
func BuildAll() error {
	for _, p := range registered {
		if err := p.Build(); err != nil {
			return fmt.Errorf("%s build failed: %w", p.Name(), err)
		}
	}
	return nil
}

// NeedsRebuild reports whether any plugin requires a rebuild for the given path.
func NeedsRebuild(path string) bool {
	for _, p := range registered {
		if p.ShouldRebuild(path) {
			return true
		}
	}
	return false
}
