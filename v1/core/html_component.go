//go:build js && wasm

package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/rfwlab/rfw/v1/dom"
	hostclient "github.com/rfwlab/rfw/v1/hostclient"
	js "github.com/rfwlab/rfw/v1/js"
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
	Props             map[string]any
	Slots             map[string]any
	HostComponent     string
	conditionContents map[string]ConditionContent
	hostVars          []string
	hostCmds          []string
	component         Component
	onMount           func(*HTMLComponent)
	onUnmount         func(*HTMLComponent)
}

func NewHTMLComponent(name string, templateFs []byte, props map[string]any) *HTMLComponent {
	id := generateComponentID(name, props)
	c := &HTMLComponent{
		ID:                id,
		Name:              name,
		TemplateFS:        templateFs,
		Dependencies:      make(map[string]Component),
		Props:             props,
		Slots:             make(map[string]any),
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
	dom.RegisterBindings(c.ID, c.Name, template)

	if store != nil {
		c.Store = store
	} else {
		c.Store = state.GlobalStoreManager.GetStore("app", "default")
		if c.Store == nil {
			panic(fmt.Sprintf("No store provided and no default store found for component %s", c.Name))
		}
	}
}

func (c *HTMLComponent) Render() (renderedTemplate string) {
	defer func() {
		if r := recover(); r != nil {
			jsStack := js.Error().New().Get("stack").String()
			goStack := string(debug.Stack())
			panic(fmt.Sprintf("%v\nGo stack:\n%s\nJS stack:\n%s", r, goStack, jsStack))
		}
	}()

	c.unsubscribes.Run()

	renderedTemplate = c.Template
	renderedTemplate = strings.Replace(renderedTemplate, "<root", fmt.Sprintf("<root data-component-id=\"%s\"", c.ID), 1)

	// Extract slot contents destined for child components
	renderedTemplate = extractSlotContents(renderedTemplate, c)

	// Replace this component's slot placeholders with provided content or fallbacks
	renderedTemplate = replaceSlotPlaceholders(renderedTemplate, c)

	for key, value := range c.Props {
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, fmt.Sprintf("%v", value))
	}

	// Register @include directives that supply inline props
	renderedTemplate = replaceComponentIncludes(renderedTemplate, c)

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

	// Handle host variable and command placeholders
	if c.HostComponent != "" {
		renderedTemplate = replaceHostPlaceholders(renderedTemplate, c)
	}

	// Handle @if:condition syntax for conditional rendering
	renderedTemplate = replaceConditionals(renderedTemplate, c)

	// Handle @on:event:handler and @event:handler syntax for event binding
	renderedTemplate = replaceEventHandlers(renderedTemplate)

	// Handle rt-is="ComponentName" for dynamic component loading
	renderedTemplate = replaceRtIsAttributes(renderedTemplate, c)

	// Render any components introduced via rt-is placeholders
	renderedTemplate = replaceIncludePlaceholders(c, renderedTemplate)

	if c.HostComponent != "" {
		hostclient.RegisterComponent(c.ID, c.HostComponent, c.hostVars)
	}

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

func (c *HTMLComponent) OnMount() {
	if c.onMount != nil {
		c.onMount(c)
	}
}

func (c *HTMLComponent) OnUnmount() {
	if c.onUnmount != nil {
		c.onUnmount(c)
	}
}

func (c *HTMLComponent) SetOnMount(fn func(*HTMLComponent)) {
	c.onMount = fn
}

func (c *HTMLComponent) SetOnUnmount(fn func(*HTMLComponent)) {
	c.onUnmount = fn
}

func (c *HTMLComponent) WithLifecycle(onMount, onUnmount func(*HTMLComponent)) *HTMLComponent {
	c.onMount = onMount
	c.onUnmount = onUnmount
	return c
}

func (c *HTMLComponent) SetComponent(component Component) {
	c.component = component
}

func (c *HTMLComponent) SetSlots(slots map[string]any) {
	if c.Slots == nil {
		c.Slots = make(map[string]any)
	}
	for k, v := range slots {
		c.Slots[k] = v
	}
}

func (c *HTMLComponent) SetRouteParams(params map[string]string) {
	if c.Props == nil {
		c.Props = make(map[string]any)
	}
	for k, v := range params {
		c.Props[k] = v
	}
}

// AddHostComponent links this HTML component to a server-side HostComponent
// by name. When running in SSC mode, messages from the wasm runtime will be
// routed to the corresponding host component on the server.
func (c *HTMLComponent) AddHostComponent(name string) {
	c.HostComponent = name
}

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
