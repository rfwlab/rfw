//go:build js && wasm

package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/internal/rendertrace"
	"github.com/rfwlab/rfw/v2/state"
)

var componentSeq atomic.Uint64

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

// HTMLComponent renders RTML templates and manages their component state.
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
	hostComponents    []string
	conditionContents map[string]ConditionContent
	// condSeq numbers the conditional blocks of one render pass: two @if
	// blocks carrying the same condition would otherwise hash to the same id,
	// share one content entry and patch each other's markup.
	condSeq int
	// forSeq numbers the loops of one render pass so every loop keeps the id
	// its rows carry in the DOM and can be patched on its own
	forSeq            int
	exprContents      map[string]string
	classExprContents map[string]string
	hostVars          []string
	hostCmds          []string
	component         Component
	mounted           bool
	onMount           func(*HTMLComponent)
	onUnmount         func(*HTMLComponent)
	onParams          func(*HTMLComponent, map[string]string)
	handlers          map[string]func()
	domHooks          []dom.LifecycleHook
	domHookStops      []func()
	scope             *Scope
	parent            *HTMLComponent
	provides          map[string]any
	cache             map[string]string
	lastCacheKey      string
	metricsMu         sync.Mutex
	renderCount       int
	totalRender       time.Duration
	lastRender        time.Duration
	timeline          []ComponentTimelineEntry
}

// ComponentStats contains aggregated render metrics for an HTML component.
type ComponentStats struct {
	RenderCount   int
	TotalRender   time.Duration
	LastRender    time.Duration
	AverageRender time.Duration
	Timeline      []ComponentTimelineEntry
}

// ComponentTimelineEntry represents a point-in-time event collected for diagnostics.
type ComponentTimelineEntry struct {
	Kind      string
	Timestamp time.Time
	Duration  time.Duration
}

// NewHTMLComponent creates a component from an RTML template and initial props.
func NewHTMLComponent(name string, templateFs []byte, props map[string]any) *HTMLComponent {
	id := generateComponentID(name, props)
	c := &HTMLComponent{
		ID:                id,
		Name:              name,
		TemplateFS:        templateFs,
		Dependencies:      make(map[string]Component),
		Props:             props,
		Slots:             make(map[string]any),
		handlers:          make(map[string]func()),
		scope:             NewScope(),
		conditionContents: make(map[string]ConditionContent),
		exprContents:      make(map[string]string),
		classExprContents: make(map[string]string),
	}
	// Attempt automatic cleanup when component is garbage collected.
	runtime.SetFinalizer(c, func(hc *HTMLComponent) { hc.Unmount() })
	return c
}

// Init attaches a state store and prepares the component template.
func (c *HTMLComponent) Init(store *state.Store) {
	if c.Store != nil {
		return
	}
	template, err := LoadComponentTemplate(c.TemplateFS)
	if err != nil {
		panic(fmt.Sprintf("Error loading template for component %s: %v", c.Name, err))
	}
	template = devOverrideTemplate(c, template)
	c.Template = template
	dom.RegisterBindings(c.ID, c.Name, template)
	devRegisterComponent(c)

	if store != nil {
		c.Store = store
	} else {
		c.Store = state.GlobalStoreManager.GetStore("app", "default")
		if c.Store == nil {
			c.Store = state.NewStore("default", state.WithModule("app"))
		}
	}
}

// RenderFresh clears the render cache and re-renders. Reactive updates (store
// OnChange, signal effects) call this so a state change always produces
// up-to-date HTML instead of a stale cached render. Fixes the bug where a
// store.Set did not re-render @for / @expr / store-bound templates because the
// cache key hashes only Props/Dependencies, not the bound store state.
func (c *HTMLComponent) RenderFresh() string {
	c.Invalidate()
	return c.Render()
}

// Invalidate drops the render cache of this component and of everything it
// includes. The cache key covers props and dependency identity, never the
// store state a template binds to, so an included component handed back its
// first render forever: a dependency whose markup depends on a shared store
// key (an @if on a global flag) froze at the value it had when the parent
// first painted.
func (c *HTMLComponent) Invalidate() {
	c.cache = nil
	c.lastCacheKey = ""
	for _, dep := range c.Dependencies {
		if d, ok := dep.(interface{ Invalidate() }); ok {
			d.Invalidate()
		}
	}
}

// Render evaluates the component template.
func (c *HTMLComponent) Render() (renderedTemplate string) {
	start := time.Now()
	defer func() { c.recordRender(time.Since(start)) }()
	key := c.cacheKey()
	if c.cache != nil {
		if val, ok := c.cache[key]; ok {
			renderedTemplate = val
			return
		}
		if c.lastCacheKey != "" && c.lastCacheKey != key {
			delete(c.cache, c.lastCacheKey)
		}
	} else {
		c.cache = make(map[string]string)
	}
	defer func() {
		if r := recover(); r != nil {
			ReportError(r, fmt.Sprintf("Render: %s (ID: %s)", c.Name, c.ID))
			renderedTemplate = ""
		}
	}()

	c.unsubscribes.Run()

	renderedTemplate = c.Template
	renderedTemplate = strings.Replace(renderedTemplate, "<root", fmt.Sprintf("<root data-component-id=\"%s\"", c.ID), 1)

	// Extract slot contents destined for child components
	renderedTemplate = extractSlotContents(renderedTemplate, c)

	// Replace this component's slot placeholders with provided content or fallbacks
	renderedTemplate = replaceSlotPlaceholders(renderedTemplate, c)

	// {{prop}} substitutions are HTML-escaped like @prop; @rawprop remains the
	// explicit escape hatch for trusted markup.
	for key, value := range c.Props {
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, escapeValue(value))
	}

	// Register @include directives that supply inline props
	renderedTemplate = replaceComponentIncludes(renderedTemplate, c)

	// Handle @include:componentName syntax for dependencies
	renderedTemplate = replaceIncludePlaceholders(c, renderedTemplate)

	// Handle @for loops
	renderedTemplate = replaceForPlaceholders(renderedTemplate, c)

	renderedTemplate = replaceStorePlaceholders(renderedTemplate, c)
	renderedTemplate = replaceSignalPlaceholders(renderedTemplate, c)
	renderedTemplate = replaceExprInClassAttr(renderedTemplate, c)
	renderedTemplate = replaceExprPlaceholders(renderedTemplate, c)

	// Handle @prop:propName syntax for props
	renderedTemplate = replacePropPlaceholders(renderedTemplate, c)

	// Handle plugin variable and command placeholders
	renderedTemplate = replacePluginPlaceholders(renderedTemplate)

	// Handle host variable and command placeholders
	if len(c.hostComponentNames()) > 0 {
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

	// Handle constructor decorators like [ref] and [key expr]
	renderedTemplate = replaceConstructors(renderedTemplate)

	if hostRegisterComponent != nil {
		for _, name := range c.hostComponentNames() {
			hostRegisterComponent(c.ID, name, c.hostVars)
		}
	}

	c.cache[key] = renderedTemplate
	c.lastCacheKey = key
	return renderedTemplate
}

const componentTimelineLimit = 64

func (c *HTMLComponent) recordRender(duration time.Duration) {
	if c == nil {
		return
	}
	c.metricsMu.Lock()
	c.renderCount++
	c.totalRender += duration
	c.lastRender = duration
	c.appendTimelineLocked(ComponentTimelineEntry{
		Kind:      "render",
		Timestamp: time.Now(),
		Duration:  duration,
	})
	c.metricsMu.Unlock()
}

func (c *HTMLComponent) appendTimelineLocked(entry ComponentTimelineEntry) {
	if entry.Kind == "" {
		return
	}
	if c.timeline == nil {
		c.timeline = make([]ComponentTimelineEntry, 0, 8)
	}
	c.timeline = append(c.timeline, entry)
	if len(c.timeline) > componentTimelineLimit {
		c.timeline = append([]ComponentTimelineEntry(nil), c.timeline[len(c.timeline)-componentTimelineLimit:]...)
	}
}

// Stats returns a snapshot of the component's render metrics.
func (c *HTMLComponent) Stats() ComponentStats {
	c.metricsMu.Lock()
	defer c.metricsMu.Unlock()
	stats := ComponentStats{
		RenderCount: c.renderCount,
		TotalRender: c.totalRender,
		LastRender:  c.lastRender,
	}
	if c.renderCount > 0 {
		stats.AverageRender = c.totalRender / time.Duration(c.renderCount)
	}
	if len(c.timeline) > 0 {
		stats.Timeline = append(stats.Timeline, c.timeline...)
	}
	return stats
}

// AddDependency attaches a child component to a template placeholder.
func (c *HTMLComponent) AddDependency(placeholderName string, dep Component) {
	if c.Dependencies == nil {
		c.Dependencies = make(map[string]Component)
	}
	if depComp, ok := dep.(*HTMLComponent); ok {
		depComp.Init(c.Store)
		depComp.parent = c
	}
	c.Dependencies[placeholderName] = dep
}

// Unmount releases component resources and child dependencies.
func (c *HTMLComponent) Unmount() {
	// The idempotence guard keeps finalizers from repeating lifecycle cleanup.
	if !c.mounted {
		return
	}
	c.mounted = false
	cancelComponentRender(c.ID)
	if rendertrace.Enabled() {
		parentID := ""
		if c.parent != nil {
			parentID = c.parent.ID
		}
		rendertrace.Emit(rendertrace.Record{
			Event:             "unmounted",
			BatchID:           rendertrace.NextBatchID(),
			RenderID:          rendertrace.NextRenderID(),
			ComponentID:       c.ID,
			ComponentName:     c.Name,
			ParentComponentID: parentID,
			Depth:             componentDepth(c),
			Cause:             rendertrace.Cause{Kind: "explicit"},
			Outcome:           "unmounted",
		})
	}
	devUnregisterComponent(c)
	if c.component != nil {
		c.runLifecycle("OnUnmount", c.component.OnUnmount)
	}
	dom.UnmountLifecycleHooks(c.ID)
	c.releaseDOMHooks()
	if c.scope != nil {
		c.scope.Close()
	}

	dom.RemoveComponentSignals(c.ID)
	dom.ReleaseInputBindings(c.ID)
	dom.ReleaseComponentHandlers(c.ID)
	root := dom.ComponentRoot(c.ID)
	if !root.IsNull() && !root.IsUndefined() {
		dom.RemoveDelegatedEvents(c.ID, root.Value)
	}
	log.Printf("Unsubscribing %s from all stores", c.Name)
	c.unsubscribes.Run()

	for _, dep := range c.Dependencies {
		dependency := dep
		c.runLifecycle("dependency unmount", dependency.Unmount)
	}
}

// Mount activates the component and its child dependencies.
func (c *HTMLComponent) Mount() {
	c.mounted = true
	if c.scope == nil || c.scope.Closed() {
		c.scope = NewScope()
	}
	c.registerHandlers()
	c.registerDOMHooks()
	for _, dep := range c.Dependencies {
		dependency := dep
		c.runLifecycle("dependency mount", dependency.Mount)
	}
	root := dom.ComponentRoot(c.ID)
	if !root.IsNull() && !root.IsUndefined() {
		dom.DelegateEvents(c.ID, root.Value)
	}
	if c.component != nil {
		c.runLifecycle("OnMount", c.component.OnMount)
	}
	dom.MountLifecycleHooks(c.ID)
}

func (c *HTMLComponent) runLifecycle(phase string, fn func()) {
	defer func() {
		if recovered := recover(); recovered != nil {
			ReportError(recovered, phase+": "+c.Name+" (ID: "+c.ID+")")
		}
	}()
	fn()
}

// Scope returns the lifecycle scope owned by this component.
func (c *HTMLComponent) Scope() *Scope {
	if c.scope == nil || c.scope.Closed() {
		c.scope = NewScope()
	}
	return c.scope
}

// Effect registers a reactive effect that stops on unmount.
func (c *HTMLComponent) Effect(fn func() func()) {
	c.Scope().Defer(state.Effect(fn))
}

// DOMHook registers root lifecycle callbacks owned by this component.
func (c *HTMLComponent) DOMHook(hook dom.LifecycleHook) {
	c.domHooks = append(c.domHooks, hook)
	if c.mounted {
		c.domHookStops = append(c.domHookStops, dom.RegisterLifecycleHook(c.ID, hook))
		dom.MountLifecycleHooks(c.ID)
	}
}

func (c *HTMLComponent) registerDOMHooks() {
	c.releaseDOMHooks()
	for _, hook := range c.domHooks {
		c.domHookStops = append(c.domHookStops, dom.RegisterLifecycleHook(c.ID, hook))
	}
}

func (c *HTMLComponent) releaseDOMHooks() {
	for _, stop := range c.domHookStops {
		stop()
	}
	c.domHookStops = nil
}

// On registers an event handler owned by this component instance.
func (c *HTMLComponent) On(name string, fn func()) {
	if name == "" {
		panic("core.HTMLComponent.On: empty handler name")
	}
	if fn == nil {
		panic("core.HTMLComponent.On: nil fn")
	}
	c.handlers[name] = fn
	dom.RegisterComponentHandlerFunc(c.ID, name, fn)
}

func (c *HTMLComponent) registerHandlers() {
	for name, fn := range c.handlers {
		dom.RegisterComponentHandlerFunc(c.ID, name, fn)
	}
}

// GetName returns the component name.
func (c *HTMLComponent) GetName() string {
	return c.Name
}

// GetID returns the component identifier.
func (c *HTMLComponent) GetID() string {
	return c.ID
}

// GetRef returns the DOM element annotated with a matching constructor
// decorator. It searches within this component's root element using the
// data-ref attribute injected during template rendering.
func (c *HTMLComponent) GetRef(name string) dom.Element {
	root := dom.ComponentRoot(c.ID)
	if root.IsNull() || root.IsUndefined() {
		return dom.Element{}
	}
	return root.Query(fmt.Sprintf(`[data-ref="%s"]`, name))
}

// OnMount runs the configured mount callback.
func (c *HTMLComponent) OnMount() {
	if c.onMount != nil {
		c.onMount(c)
	}
}

// OnUnmount runs the configured unmount callback.
func (c *HTMLComponent) OnUnmount() {
	if c.onUnmount != nil {
		c.onUnmount(c)
	}
	c.mounted = false
}

// IsMounted reports whether the component is mounted.
func (c *HTMLComponent) IsMounted() bool {
	return c.mounted
}

// OnParams runs the configured route-parameter callback.
func (c *HTMLComponent) OnParams(params map[string]string) {
	if c.onParams != nil {
		c.onParams(c, params)
	}
}

// SetOnParams configures the route-parameter callback.
func (c *HTMLComponent) SetOnParams(fn func(*HTMLComponent, map[string]string)) {
	c.onParams = fn
}

// SetOnMount configures the mount callback.
func (c *HTMLComponent) SetOnMount(fn func(*HTMLComponent)) {
	c.onMount = fn
}

// SetOnUnmount configures the unmount callback.
func (c *HTMLComponent) SetOnUnmount(fn func(*HTMLComponent)) {
	c.onUnmount = fn
}

// WithLifecycle configures mount and unmount callbacks.
func (c *HTMLComponent) WithLifecycle(onMount, onUnmount func(*HTMLComponent)) *HTMLComponent {
	c.onMount = onMount
	c.onUnmount = onUnmount
	return c
}

// SetComponent attaches the component lifecycle implementation.
func (c *HTMLComponent) SetComponent(component Component) {
	c.component = component
}

// SetSlots merges named slot content into the component.
func (c *HTMLComponent) SetSlots(slots map[string]any) {
	if c.Slots == nil {
		c.Slots = make(map[string]any)
	}
	for k, v := range slots {
		c.Slots[k] = v
	}
}

// Provide stores a value on this component so that descendants can
// retrieve it with Inject. It creates the map on first use.
func (c *HTMLComponent) Provide(key string, val any) {
	if c.provides == nil {
		c.provides = make(map[string]any)
	}
	c.provides[key] = val
}

// Inject searches for a provided value starting from this component and
// walking up the parent chain. It returns the value as `any` and whether it
// was found. Callers can type-assert the result.
func (c *HTMLComponent) Inject(key string) (any, bool) {
	if c.provides != nil {
		if v, ok := c.provides[key]; ok {
			return v, true
		}
	}
	if c.parent != nil {
		return c.parent.Inject(key)
	}
	return nil, false
}

// Inject performs a typed lookup of a provided component value.
func Inject[T any](c *HTMLComponent, key string) (T, bool) {
	v, ok := c.Inject(key)
	if !ok {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	return t, ok
}

// SetRouteParams merges route parameters into component props.
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
// routed to the corresponding host component on the server. It may be called
// multiple times (e.g. a composition struct with several host fields): every
// name is registered, and HostComponent keeps the first one as the primary.
func (c *HTMLComponent) AddHostComponent(name string) {
	for _, n := range c.hostComponents {
		if n == name {
			return
		}
	}
	c.hostComponents = append(c.hostComponents, name)
	if c.HostComponent == "" {
		c.HostComponent = name
	}
}

// hostComponentNames returns every host component linked to this component,
// including a HostComponent assigned directly to the exported field.
func (c *HTMLComponent) hostComponentNames() []string {
	if len(c.hostComponents) > 0 {
		return c.hostComponents
	}
	if c.HostComponent != "" {
		return []string{c.HostComponent}
	}
	return nil
}

func (c *HTMLComponent) cacheKey() string {
	hasher := sha256.New()
	hasher.Write([]byte(serializeProps(c.Props)))

	if len(c.Dependencies) > 0 {
		deps := make([]string, 0, len(c.Dependencies))
		for name, dep := range c.Dependencies {
			deps = append(deps, name+dep.GetID())
		}
		sort.Strings(deps)
		for _, d := range deps {
			hasher.Write([]byte(d))
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)[:20])
}

func generateComponentID(name string, props map[string]any) string {
	hasher := sha256.New()
	hasher.Write([]byte(name))
	propsString := serializeProps(props)
	hasher.Write([]byte(propsString))
	hasher.Write([]byte(strconv.FormatUint(componentSeq.Add(1), 10)))

	return hex.EncodeToString(hasher.Sum(nil)[:20])
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
		fmt.Fprintf(&sb, "%s=%v;", k, v)
	}

	return sb.String()
}
