// Package foundation integrates go-foundation primitives into rfw v2.
// It provides DI, type-safe events, lifecycle hooks, effect pipelines,
// and structured options for the composition system.
package foundation

import (
	"context"
	"fmt"
	"reflect"

	fndi "github.com/mirkobrombin/go-foundation/pkg/di"
	fnevents "github.com/mirkobrombin/go-foundation/pkg/events"
	fnhooks "github.com/mirkobrombin/go-foundation/pkg/hooks"
	fnpipeline "github.com/mirkobrombin/go-foundation/pkg/pipeline"
	fnresult "github.com/mirkobrombin/go-foundation/pkg/result"
)

// ── DI ──────────────────────────────────────────────────────────────────────

// Container wraps the foundation DI container with component-scoping.
type Container struct {
	*fndi.Container
}

// NewContainer creates a root DI container for the application.
func NewContainer() *Container {
	return &Container{fndi.New()}
}

// Scope creates a child container bound to a component lifecycle.
func (c *Container) Scope() *Container {
	return &Container{Container: c.Container.Scope()}
}

// ── Events ──────────────────────────────────────────────────────────────────

// EventBus wraps the foundation type-safe event bus.
type EventBus struct {
	bus *fnevents.Bus
}

// DefaultBus is the framework-wide event bus instance.
var DefaultBus = &EventBus{bus: fnevents.New()}

// Subscribe registers a handler for a typed event.
func Subscribe[T any](bus *EventBus, fn fnevents.Handler[T], priority ...fnevents.Priority) {
	fnevents.Subscribe[T](bus.bus, fn, priority...)
}

// Emit sends a typed event synchronously.
func Emit[T any](bus *EventBus, event T) error {
	return fnevents.Emit(context.Background(), bus.bus, event)
}

// EmitAsync sends a typed event without waiting for handlers.
func EmitAsync[T any](bus *EventBus, event T) {
	fnevents.EmitAsync(context.Background(), bus.bus, event)
}

// ── Hooks ───────────────────────────────────────────────────────────────────

// Lifecycle exposes foundation hooks for component lifecycle.
type Lifecycle struct {
	runner *fnhooks.Runner
}

// NewLifecycle creates a lifecycle hook runner backed by reflection.
func NewLifecycle() *Lifecycle {
	return &Lifecycle{runner: fnhooks.NewRunner()}
}

// Before registers a pre-action hook (e.g., BeforeMount, BeforeUpdate).
func (l *Lifecycle) Before(key string, fn fnhooks.HookFunc) {
	l.runner.Before(key, fn)
}

// After registers a post-action hook (e.g., AfterMount, AfterUpdate).
func (l *Lifecycle) After(key string, fn fnhooks.HookFunc) {
	l.runner.After(key, fn)
}

// Run executes an action surrounded by registered hooks.
func (l *Lifecycle) Run(ctx context.Context, key string, action func() error, args ...any) error {
	return l.runner.Run(ctx, key, action, args...)
}

// Discover scans obj via reflection for methods prefixed with prefix and
// auto-registers them as hooks. Not yet wired to foundation hooks discovery.
func (l *Lifecycle) Discover(obj any, prefix string) *Lifecycle {
	// placeholder: will wire to foundation hooks.Discovery when API stabilizes
	return l
}

// ── Pipeline ────────────────────────────────────────────────────────────────

// EffectPipeline chains middleware over signal-driven side effects.
type EffectPipeline struct {
	pip *fnpipeline.Pipeline[EffectInput, struct{}]
}

// EffectInput carries what changed and the component ID.
type EffectInput struct {
	ComponentID string
	SignalName  string
	Value       any
}

// NewEffectPipeline builds a pipeline with a no-op default handler.
func NewEffectPipeline() *EffectPipeline {
	p := fnpipeline.New[EffectInput, struct{}]()
	// default handler does nothing
	p.Then(func(_ context.Context, in EffectInput) (struct{}, error) {
		return struct{}{}, nil
	})
	return &EffectPipeline{pip: p}
}

// Use registers middleware. Middleware receives EffectInput and can call next.
func (ep *EffectPipeline) Use(mw fnpipeline.Middleware[EffectInput, struct{}]) {
	if ep.pip == nil {
		return
	}
	ep.pip.Use(mw)
}

// Process executes the pipeline for a signal change.
func (ep *EffectPipeline) Process(ctx context.Context, in EffectInput) {
	_, _ = ep.pip.Process(ctx, in)
}

// ── Result ──────────────────────────────────────────────────────────────────

// Result re-exports foundation's Result monad for async UI ops.
type Result[T any] = fnresult.Result[T]

func Ok[T any](v T) Result[T]      { return fnresult.Ok[T](v) }
func Err[T any](e error) Result[T] { return fnresult.Err[T](e) }

// ── Options ─────────────────────────────────────────────────────────────────

// Option is a functional option for composition configuration.
type Option[T any] func(o *T)

// Apply runs opts against target.
func Apply[T any](target *T, opts ...Option[T]) {
	for _, o := range opts {
		if o != nil {
			o(target)
		}
	}
}

// ── Tag Scanner ───────────────────────────────────────────────────────────────

// TagScanner reads struct field tags for the composition system.
// It is stateless and safe for concurrent use.
var TagScanner = &tagScanner{}

type tagScanner struct{}

// Meta holds all rfw tag metadata extracted from a single struct type.
type Meta struct {
	Signals   []SignalMeta
	Stores    []StoreMeta
	Props     []PropMeta
	Refs      []string
	Hosts     []HostMeta
	Events    []EventMeta
	Injects   []InjectMeta
	FSMs      []FSMMeta
	Histories []HistoryMeta
}

// Field-level metadata structs.
type SignalMeta struct {
	Field reflect.StructField
	Name  string // defaults to field name
}
type StoreMeta struct {
	Field reflect.StructField
	Name  string // from tag value
}
type PropMeta struct {
	Field      reflect.StructField
	Name       string
	DefaultVal any
}
type HostMeta struct {
	Field reflect.StructField
	Name  string
}
type EventMeta struct {
	Field     reflect.StructField
	DOMEvent  string
	Handler   string
	Modifiers []string
}
type InjectMeta struct {
	Field reflect.StructField
	Key   string
}
type FSMMeta struct {
	Field      reflect.StructField
	Definition string // raw tag value
}
type HistoryMeta struct {
	Field   reflect.StructField
	Store   string
	UndoEvt string
	RedoEvt string
}

// Scan extracts Meta from the given struct (pointer to struct).
func (ts *tagScanner) Scan(v any) (*Meta, error) {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("rfw tag scanner: expected struct, got %s", typ.Kind())
	}

	m := &Meta{}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag, ok := field.Tag.Lookup("rfw")
		if !ok {
			continue
		}
		if tag == "" {
			continue
		}
		ts.parseTag(field, tag, m)
	}
	return m, nil
}

func (ts *tagScanner) parseTag(field reflect.StructField, tag string, m *Meta) {
	// Shorthand with empty value means the directive itself is the marker.
	switch tag {
	case "signal":
		m.Signals = append(m.Signals, SignalMeta{Field: field, Name: field.Name})
		return
	case "prop":
		m.Props = append(m.Props, PropMeta{Field: field, Name: field.Name})
		return
	case "ref":
		m.Refs = append(m.Refs, field.Name)
		return
	case "inject":
		m.Injects = append(m.Injects, InjectMeta{Field: field, Key: field.Name})
		return
	}

	// key:value pairs
	parts := splitTag(tag)
	switch parts[0] {
	case "store":
		if len(parts) > 1 {
			m.Stores = append(m.Stores, StoreMeta{Field: field, Name: parts[1]})
		}
	case "host":
		if len(parts) > 1 {
			m.Hosts = append(m.Hosts, HostMeta{Field: field, Name: parts[1]})
		}
	case "event":
		if len(parts) >= 3 {
			modifiers := []string{}
			if len(parts) > 3 {
				modifiers = parts[3:]
			}
			m.Events = append(m.Events, EventMeta{
				Field: field, DOMEvent: parts[1], Handler: parts[2], Modifiers: modifiers,
			})
		}
	case "inject":
		key := field.Name
		if len(parts) > 1 {
			key = parts[1]
		}
		m.Injects = append(m.Injects, InjectMeta{Field: field, Key: key})
	case "fsm":
		if len(parts) > 1 {
			m.FSMs = append(m.FSMs, FSMMeta{Field: field, Definition: parts[1]})
		}
	case "history":
		if len(parts) >= 4 {
			m.Histories = append(m.Histories, HistoryMeta{
				Field: field, Store: parts[1], UndoEvt: parts[2], RedoEvt: parts[3],
			})
		}
	}
}

// splitTag splits a colon-separated tag safely.
func splitTag(tag string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(tag); i++ {
		if tag[i] == ':' {
			parts = append(parts, tag[start:i])
			start = i + 1
		}
	}
	parts = append(parts, tag[start:])
	return parts
}
