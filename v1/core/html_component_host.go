//go:build !js || !wasm

package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/rfwlab/rfw/v1/rtml"
)

// HTMLComponent is a lightweight server-side component capable of rendering RTML templates.
type HTMLComponent struct {
	Name         string
	TemplateFS   []byte
	Props        map[string]any
	Slots        map[string]any
	ID           string
	component    Component
	Dependencies map[string]Component
}

// NewHTMLComponent initializes a server-side HTMLComponent.
func NewHTMLComponent(name string, templateFS []byte, props map[string]any) *HTMLComponent {
	c := &HTMLComponent{
		Name:         name,
		TemplateFS:   templateFS,
		Props:        props,
		Dependencies: make(map[string]Component),
	}
	c.ID = generateComponentID(name, props)
	return c
}

// Render executes the RTML pipeline and returns HTML with the hydration marker.
func (c *HTMLComponent) Render() (renderedTemplate string) {
	renderedTemplate, _ = LoadComponentTemplate(c.TemplateFS)
	renderedTemplate = strings.Replace(renderedTemplate, "<root", fmt.Sprintf("<root data-component-id=\"%s\"", c.ID), 1)
	ctx := rtml.Context{Props: c.Props, Slots: c.Slots, Dependencies: make(map[string]rtml.Dependency)}
	for k, dep := range c.Dependencies {
		ctx.Dependencies[k] = dep
	}
	renderedTemplate = rtml.Replace(renderedTemplate, ctx)
	return renderedTemplate
}

// SetComponent assigns the underlying component implementation.
func (c *HTMLComponent) SetComponent(comp Component) {
	c.component = comp
}

// GetName returns the component name.
func (c *HTMLComponent) GetName() string { return c.Name }

// GetID returns the component id used for hydration.
func (c *HTMLComponent) GetID() string { return c.ID }

// SetSlots stores slots used during rendering.
func (c *HTMLComponent) SetSlots(slots map[string]any) { c.Slots = slots }

// AddDependency registers a child component rendered inside this component.
func (c *HTMLComponent) AddDependency(placeholderName string, dep Component) {
	if c.Dependencies == nil {
		c.Dependencies = make(map[string]Component)
	}
	c.Dependencies[placeholderName] = dep
}

// helper functions reused from client implementation
func generateComponentID(name string, props map[string]any) string {
	hasher := sha1.New()
	hasher.Write([]byte(name))
	propsString := serializeProps(props)
	hasher.Write([]byte(propsString))
	return hex.EncodeToString(hasher.Sum(nil))
}

func serializeProps(props map[string]any) string {
	if props == nil {
		return ""
	}
	var sb strings.Builder
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := props[k]
		sb.WriteString(fmt.Sprintf("%s=%v;", k, v))
	}
	return sb.String()
}
