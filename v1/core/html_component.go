//go:build js && wasm

package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strings"

	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

type unsubscribes struct {
	funcs []func()
}

func (u *unsubscribes) Add(fn func()) { u.funcs = append(u.funcs, fn) }

func (u *unsubscribes) Run() {
	for _, fn := range u.funcs {
		fn()
	}
	u.funcs = nil
}

type HTMLComponent struct {
	ID                string
	Name              string
	Template          string
	TemplateFS        []byte
	Dependencies      map[string]Component
	unsubscribes      unsubscribes
	Store             *state.Store
	Props             map[string]interface{}
	Slots             map[string]string
	conditionContents map[string]ConditionContent
	component         Component
}

func NewHTMLComponent(name string, templateFs []byte, props map[string]interface{}) *HTMLComponent {
	id := generateComponentID(name, props)
	c := &HTMLComponent{
		ID:                id,
		Name:              name,
		TemplateFS:        templateFs,
		Dependencies:      make(map[string]Component),
		Props:             props,
		Slots:             make(map[string]string),
		conditionContents: make(map[string]ConditionContent),
	}
	// Attempt automatic cleanup when component is garbage collected.
	runtime.SetFinalizer(c, func(hc *HTMLComponent) { hc.Unmount() })
	return c
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
		c.Store = state.GlobalStoreManager.GetStore("app", "default")
		if c.Store == nil {
			panic(fmt.Sprintf("No store provided and no default store found for component %s", c.Name))
		}
	}
}

func (c *HTMLComponent) Render() string {
	c.unsubscribes.Run()

	renderedTemplate := c.Template
	renderedTemplate = strings.Replace(renderedTemplate, "<root", fmt.Sprintf("<root data-component-id=\"%s\"", c.ID), 1)

	// Extract slot contents destined for child components
	renderedTemplate = extractSlotContents(renderedTemplate, c)

	// Replace this component's slot placeholders with provided content or fallbacks
	renderedTemplate = replaceSlotPlaceholders(renderedTemplate, c)

	for key, value := range c.Props {
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, fmt.Sprintf("%v", value))
	}

	// Handle @include:componentName syntax for dependencies
	renderedTemplate = replaceIncludePlaceholders(c, renderedTemplate)

	// Handle @for loops and legacy @foreach syntax
	renderedTemplate = replaceForPlaceholders(renderedTemplate, c)
	renderedTemplate = replaceForeachPlaceholders(renderedTemplate, c)

	// Handle @store:module.storeName.varName syntax:
	// - :w stands for writeable inputs
	// - :r stands for read-only inputs (default, not required, actually not even implemented)
	renderedTemplate = replaceStorePlaceholders(renderedTemplate, c)

	// Handle @prop:propName syntax for props
	renderedTemplate = replacePropPlaceholders(renderedTemplate, c)

	// Handle @if:condition syntax for conditional rendering
	renderedTemplate = replaceConditionals(renderedTemplate, c)

	// Handle @on:event="handler" syntax for event binding
	renderedTemplate = replaceEventHandlers(renderedTemplate)

	// Handle rt-is="ComponentName" for dynamic component loading
	renderedTemplate = replaceRtIsAttributes(renderedTemplate, c)

	// Render any components introduced via rt-is placeholders
	renderedTemplate = replaceIncludePlaceholders(c, renderedTemplate)

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
	log.Printf("Unsubscribing %s from all stores", c.Name)
	c.unsubscribes.Run()

	for _, dep := range c.Dependencies {
		dep.Unmount()
	}

	if c.component != nil {
		c.component.OnUnmount()
	}
}

func (c *HTMLComponent) Mount() {
	for _, dep := range c.Dependencies {
		dep.Mount()
	}
	if c.component != nil {
		c.component.OnMount()
	}
}

func (c *HTMLComponent) GetName() string {
	return c.Name
}

func (c *HTMLComponent) GetID() string {
	return c.ID
}

func (c *HTMLComponent) OnMount() {}

func (c *HTMLComponent) OnUnmount() {}

func (c *HTMLComponent) SetComponent(component Component) {
	c.component = component
}

func (c *HTMLComponent) SetSlots(slots map[string]string) {
	if c.Slots == nil {
		c.Slots = make(map[string]string)
	}
	for k, v := range slots {
		c.Slots[k] = v
	}
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
