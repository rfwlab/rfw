package state

import (
	"encoding/json"
	"sync"
	"sync/atomic"
)

// effect represents a reactive computation registered via Effect.
type effect struct {
	run func() func()

	mu      sync.Mutex
	deps    []subscriber
	cleanup func()
}

type subscriber interface {
	remove(*effect)
}

// currentEffect tracks the effect whose run function is executing so that
// Signal.Get can register dependencies. It is an atomic pointer so background
// goroutines that Set/Get signal values do not race with effect execution,
// but dependency tracking itself assumes effects never run in parallel:
// effects re-run on the goroutine that calls Set, and Effect registration is
// expected to happen on the render goroutine.
var currentEffect atomic.Pointer[effect]

// Subscription represents a cancellable listener returned by OnChange.
type Subscription struct {
	cancel func()
	once   sync.Once
}

// Stop removes the listener and releases the associated channel.
func (s *Subscription) Stop() {
	s.once.Do(s.cancel)
}

// Signal holds a value of type T and tracks which effects depend on it.
// Get/Set are safe for concurrent use, so background goroutines may feed a
// signal; dependent effects run synchronously on the goroutine calling Set.
type Signal[T any] struct {
	mu    sync.Mutex
	value T
	subs  map[*effect]struct{}

	onChangeMu sync.Mutex
	onChange   []func(T)
	ch         chan T
	chCreated  bool
}

// NewSignal creates a new Signal with the given initial value.
func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{value: initial, subs: make(map[*effect]struct{})}
}

// Get returns the current value of the signal and registers the calling effect.
func (s *Signal[T]) Get() T {
	if s == nil {
		var zero T
		return zero
	}
	if eff := currentEffect.Load(); eff != nil {
		s.mu.Lock()
		if s.subs == nil {
			s.subs = make(map[*effect]struct{})
		}
		s.subs[eff] = struct{}{}
		s.mu.Unlock()
		eff.mu.Lock()
		eff.deps = append(eff.deps, s)
		eff.mu.Unlock()
	}
	s.mu.Lock()
	v := s.value
	s.mu.Unlock()
	return v
}

// Read implements a generic getter for use without knowing T.
func (s *Signal[T]) Read() any {
	if s == nil {
		return nil
	}
	return s.Get()
}

// SetFromHost sets the signal value from an untyped host payload.
// JSON decodes numbers as float64 and composites as []any/map[string]any,
// so payloads that do not assert directly to T are converted through a
// JSON round-trip into T. Payloads that cannot represent T are ignored.
func (s *Signal[T]) SetFromHost(raw any) {
	if s == nil {
		return
	}
	if v, ok := raw.(T); ok {
		s.Set(v)
		return
	}
	if val, ok := raw.(float64); ok {
		// Fast paths for the common numeric targets. The conversion happens
		// on a local value so it never touches s.value outside the lock.
		var conv T
		switch p := any(&conv).(type) {
		case *int:
			*p = int(val)
		case *int8:
			*p = int8(val)
		case *int16:
			*p = int16(val)
		case *int32:
			*p = int32(val)
		case *int64:
			*p = int64(val)
		case *uint:
			*p = uint(val)
		case *uint8:
			*p = uint8(val)
		case *uint16:
			*p = uint16(val)
		case *uint32:
			*p = uint32(val)
		case *uint64:
			*p = uint64(val)
		case *float32:
			*p = float32(val)
		case *float64:
			*p = val
		default:
			s.setViaJSON(raw)
			return
		}
		s.Set(conv)
		return
	}
	s.setViaJSON(raw)
}

// setViaJSON converts an arbitrary decoded payload into T by re-encoding it,
// covering structs, slices and maps that arrive as map[string]any/[]any.
func (s *Signal[T]) setViaJSON(raw any) {
	blob, err := json.Marshal(raw)
	if err != nil {
		return
	}
	var v T
	if err := json.Unmarshal(blob, &v); err != nil {
		return
	}
	s.Set(v)
}

// Set updates the signal's value and notifies dependent effects. Effects run
// synchronously on the calling goroutine, outside the signal's lock.
func (s *Signal[T]) Set(v T) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.value = v
	snapshot := make([]*effect, 0, len(s.subs))
	for eff := range s.subs {
		snapshot = append(snapshot, eff)
	}
	s.mu.Unlock()
	for _, eff := range snapshot {
		eff.runEffect()
	}
	s.notifyOnChange(v)
}

// OnChange registers a callback that fires whenever the signal's value changes.
// Returns a Subscription that can be stopped to remove the listener.
func (s *Signal[T]) OnChange(fn func(T)) *Subscription {
	if s == nil {
		return &Subscription{cancel: func() {}}
	}
	s.onChangeMu.Lock()
	idx := len(s.onChange)
	s.onChange = append(s.onChange, fn)
	s.onChangeMu.Unlock()

	sub := &Subscription{
		cancel: func() {
			s.onChangeMu.Lock()
			defer s.onChangeMu.Unlock()
			if idx < len(s.onChange) {
				s.onChange[idx] = nil
			}
		},
	}
	return sub
}

// Channel returns a read-only channel that receives the new value on each Set.
// The channel is created lazily on first call and shared across all listeners.
// It is closed automatically when all OnChange listeners are removed.
func (s *Signal[T]) Channel() <-chan T {
	if s == nil {
		return nil
	}
	s.onChangeMu.Lock()
	defer s.onChangeMu.Unlock()
	if !s.chCreated {
		s.ch = make(chan T, 8)
		s.chCreated = true
	}
	return s.ch
}

func (s *Signal[T]) notifyOnChange(v T) {
	s.onChangeMu.Lock()
	listeners := make([]func(T), len(s.onChange))
	copy(listeners, s.onChange)
	ch := s.ch
	hasCh := s.chCreated
	s.onChangeMu.Unlock()

	for _, fn := range listeners {
		if fn != nil {
			fn(v)
		}
	}
	if hasCh && ch != nil {
		select {
		case ch <- v:
		default:
		}
	}
}

func (s *Signal[T]) SubCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subs)
}

func (s *Signal[T]) remove(e *effect) {
	s.mu.Lock()
	if s.subs != nil {
		delete(s.subs, e)
	}
	s.mu.Unlock()
}

// detach runs the pending cleanup and unsubscribes the effect from all of its
// dependencies, leaving it ready to re-track (runEffect) or stopped for good.
func (e *effect) detach() {
	e.mu.Lock()
	cleanup := e.cleanup
	e.cleanup = nil
	deps := e.deps
	e.deps = nil
	e.mu.Unlock()
	if cleanup != nil {
		cleanup()
	}
	for _, dep := range deps {
		dep.remove(e)
	}
}

func (e *effect) runEffect() {
	e.detach()
	prev := currentEffect.Load()
	currentEffect.Store(e)
	cleanup := e.run()
	currentEffect.Store(prev)
	e.mu.Lock()
	e.cleanup = cleanup
	e.mu.Unlock()
}

func (e *effect) stop() {
	e.detach()
}

// Effect registers a reactive computation that automatically re-runs when its
// dependent signals change. The provided function may return a cleanup function
// that will run before the next execution and when the effect is stopped.
func Effect(fn func() func()) func() {
	e := &effect{run: fn}
	e.runEffect()
	return e.stop
}
