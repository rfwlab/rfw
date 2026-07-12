package state

import (
	"encoding/json"
	"sync"
)

// effect represents a reactive computation registered via Effect.
type effect struct {
	run     func() func()
	deps    []subscriber
	cleanup func()
}

type subscriber interface {
	remove(*effect)
}

var (
	currentEffect *effect
)

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
type Signal[T any] struct {
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
	if currentEffect != nil {
		if s.subs == nil {
			s.subs = make(map[*effect]struct{})
		}
		s.subs[currentEffect] = struct{}{}
		currentEffect.deps = append(currentEffect.deps, s)
	}
	return s.value
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
		// Fast paths for the common numeric targets.
		switch p := any(&s.value).(type) {
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
		s.Set(s.value)
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

// Set updates the signal's value and notifies dependent effects.
func (s *Signal[T]) Set(v T) {
	if s == nil {
		return
	}
	s.value = v
	if s.subs == nil {
		s.subs = make(map[*effect]struct{})
	}
	snapshot := make(map[*effect]struct{}, len(s.subs))
	for eff := range s.subs {
		snapshot[eff] = struct{}{}
	}
	for eff := range snapshot {
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
	if s.subs == nil {
		return 0
	}
	return len(s.subs)
}

func (s *Signal[T]) remove(e *effect) {
	if s.subs != nil {
		delete(s.subs, e)
	}
}

func (e *effect) runEffect() {
	if e.cleanup != nil {
		e.cleanup()
		e.cleanup = nil
	}
	for _, dep := range e.deps {
		dep.remove(e)
	}
	e.deps = nil
	prev := currentEffect
	currentEffect = e
	e.cleanup = e.run()
	currentEffect = prev
}

func (e *effect) stop() {
	if e.cleanup != nil {
		e.cleanup()
		e.cleanup = nil
	}
	for _, dep := range e.deps {
		dep.remove(e)
	}
	e.deps = nil
}

// Effect registers a reactive computation that automatically re-runs when its
// dependent signals change. The provided function may return a cleanup function
// that will run before the next execution and when the effect is stopped.
func Effect(fn func() func()) func() {
	e := &effect{run: fn}
	e.runEffect()
	return e.stop
}
