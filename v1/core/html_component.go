//go:build js && wasm

package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

type HTMLComponent struct {
	ID                string
	Name              string
	Template          string
	TemplateFS        []byte
	Dependencies      map[string]Component
	unsubscribes      []func()
	Store             *state.Store
	Props             map[string]interface{}
	conditionContents map[string]ConditionContent
}

func NewHTMLComponent(name string, templateFs []byte, props map[string]interface{}) *HTMLComponent {
	id := generateComponentID(name, props)
	return &HTMLComponent{
		ID:                id,
		Name:              name,
		TemplateFS:        templateFs,
		Dependencies:      make(map[string]Component),
		Props:             props,
		conditionContents: make(map[string]ConditionContent),
	}
}

func (c *HTMLComponent) Init(store *state.Store) {
	template, err := LoadComponentTemplate(c.TemplateFS)
	if err != nil {
		panic(fmt.Sprintf("Error loading template for component %s: %v", c.Name, err))
	}
	c.Template = template

	if store != nil {
		c.Store = store
	} else {
		c.Store = state.GlobalStoreManager.GetStore("default")
		if c.Store == nil {
			panic(fmt.Sprintf("No store provided and no default store found for component %s", c.Name))
		}
	}
}

func (c *HTMLComponent) Render() string {
	for _, unsubscribe := range c.unsubscribes {
		unsubscribe()
	}
	c.unsubscribes = nil

	renderedTemplate := c.Template
	renderedTemplate = strings.Replace(renderedTemplate, "<root", fmt.Sprintf("<root data-component-id=\"%s\"", c.ID), 1)

	for key, value := range c.Props {
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, fmt.Sprintf("%v", value))
	}

	// Handle @include:componentName syntax for dependencies
	renderedTemplate = replaceIncludePlaceholders(c, renderedTemplate)

	// Handle @foreach:collection as item syntax
	renderedTemplate = replaceForeachPlaceholders(renderedTemplate, c)

	// Handle @store:storeName.varName syntax:
	// - :w stands for writeable inputs
	// - :r stands for read-only inputs (default, not required, actually not even implemented)
	renderedTemplate = replaceStorePlaceholders(renderedTemplate, c)

	// Handle @prop:propName syntax for props
	renderedTemplate = replacePropPlaceholders(renderedTemplate, c)

	// Handle @if:condition syntax for conditional rendering
	renderedTemplate = replaceConditionals(renderedTemplate, c)

	// Handle @on:event="handler" syntax for event binding
	renderedTemplate = replaceEventHandlers(renderedTemplate)

	return renderedTemplate
}

func (c *HTMLComponent) AddDependency(placeholderName string, dep Component) {
	if c.Dependencies == nil {
		c.Dependencies = make(map[string]Component)
	}
	if depComp, ok := dep.(*HTMLComponent); ok {
		depComp.Init(c.Store)
	}
	c.Dependencies[placeholderName] = dep
}

func (c *HTMLComponent) Unmount() {
	dom.RemoveEventListeners(c.ID)

	for _, unsubscribe := range c.unsubscribes {
		log.Printf("Unsubscribing %s from all stores", c.Name)
		unsubscribe()
	}
	c.unsubscribes = nil

	for _, dep := range c.Dependencies {
		dep.Unmount()
	}
}

func (c *HTMLComponent) Mount() {
	for _, dep := range c.Dependencies {
		dep.Mount()
	}
}

func (c *HTMLComponent) GetName() string {
	return c.Name
}

func (c *HTMLComponent) GetID() string {
	return c.ID
}

func (c *HTMLComponent) SetRouteParams(params map[string]string) {
	if c.Props == nil {
		c.Props = make(map[string]interface{})
	}
	for k, v := range params {
		c.Props[k] = v
	}
}

func generateComponentID(name string, props map[string]interface{}) string {
	hasher := sha1.New()
	hasher.Write([]byte(name))
	propsString := serializeProps(props)
	hasher.Write([]byte(propsString))

	return hex.EncodeToString(hasher.Sum(nil))
}

func serializeProps(props map[string]interface{}) string {
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
