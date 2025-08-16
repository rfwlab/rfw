package state

type ReactiveVar struct {
	value     string
	listeners []func(string)
}

func NewReactiveVar(initial string) *ReactiveVar {
	return &ReactiveVar{
		value: initial,
	}
}

func (rv *ReactiveVar) Set(newValue string) {
	rv.value = newValue
	for _, listener := range rv.listeners {
		listener(newValue)
	}
}

func (rv *ReactiveVar) Get() string {
	return rv.value
}

func (rv *ReactiveVar) OnChange(listener func(string)) {
	rv.listeners = append(rv.listeners, listener)
}

// Computed represents a derived state value based on other store keys.
// It holds the target key for the computed value, the list of dependencies
// and the function used to calculate the value.
type Computed struct {
	key     string
	deps    []string
	compute func(map[string]interface{}) interface{}
}

// NewComputed creates a new Computed value.
func NewComputed(key string, deps []string, compute func(map[string]interface{}) interface{}) *Computed {
	return &Computed{key: key, deps: deps, compute: compute}
}

// Key returns the store key associated with the computed value.
func (c *Computed) Key() string { return c.key }

// Deps returns the list of keys this computed value depends on.
func (c *Computed) Deps() []string { return c.deps }

// Evaluate executes the compute function using the provided state and returns
// the result.
func (c *Computed) Evaluate(state map[string]interface{}) interface{} {
	return c.compute(state)
}

// Watcher represents a callback that reacts to changes on specific store keys.
// When any of the dependencies change, the associated function is triggered.
type Watcher struct {
	deps []string
	run  func(map[string]interface{})
}

// NewWatcher creates a new Watcher.
func NewWatcher(deps []string, run func(map[string]interface{})) *Watcher {
	return &Watcher{deps: deps, run: run}
}

// Deps returns the list of keys the watcher observes.
func (w *Watcher) Deps() []string { return w.deps }

// Run triggers the watcher with the provided state.
func (w *Watcher) Run(state map[string]interface{}) { w.run(state) }
