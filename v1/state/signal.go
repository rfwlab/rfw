package state

// effect represents a reactive computation registered via Effect.
type effect struct {
	run     func() func()
	deps    []subscriber
	cleanup func()
}

type subscriber interface {
	remove(*effect)
}

var currentEffect *effect

// Signal holds a value of type T and tracks which effects depend on it.
type Signal[T any] struct {
	value T
	subs  map[*effect]struct{}
}

// NewSignal creates a new Signal with the given initial value.
func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{value: initial, subs: make(map[*effect]struct{})}
}

// Get returns the current value of the signal and registers the calling effect.
func (s *Signal[T]) Get() T {
	if currentEffect != nil {
		s.subs[currentEffect] = struct{}{}
		currentEffect.deps = append(currentEffect.deps, s)
	}
	return s.value
}

// Read implements a generic getter for use without knowing T.
func (s *Signal[T]) Read() any { return s.Get() }

// Set updates the signal's value and notifies dependent effects.
func (s *Signal[T]) Set(v T) {
	s.value = v
	for eff := range s.subs {
		eff.runEffect()
	}
}

func (s *Signal[T]) remove(e *effect) { delete(s.subs, e) }

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
