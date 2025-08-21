//go:build !js || !wasm

package core

// Component defines the minimal interface for SSR-compatible components.
type Component interface {
	Render() string
	GetName() string
	GetID() string
}

// ComponentRegistry holds constructors for components available to the SSR renderer.
var ComponentRegistry = map[string]func() Component{}

// RegisterComponent registers a component constructor for lookup by name.
func RegisterComponent(name string, constructor func() Component) {
	ComponentRegistry[name] = constructor
}

// NewComponent returns a basic HTMLComponent initialized with the provided template and props.
func NewComponent(name string, templateFS []byte, props map[string]any) *HTMLComponent {
	c := NewHTMLComponent(name, templateFS, props)
	c.SetComponent(c)
	return c
}

// NewComponentWith binds an existing component implementation to the HTMLComponent.
func NewComponentWith[T Component](name string, templateFS []byte, props map[string]any, self T) *HTMLComponent {
	c := NewHTMLComponent(name, templateFS, props)
	if any(self) != nil {
		c.SetComponent(self)
	} else {
		c.SetComponent(c)
	}
	return c
}
