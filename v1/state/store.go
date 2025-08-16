package state

import (
	"fmt"
	"reflect"
)

// StoreHook, if non-nil, is invoked on every mutation allowing external
// observers (e.g. plugins) to react to state changes without creating an
// import cycle with core.
var StoreHook func(module, store, key string, value interface{})

// StoreOption configures optional behaviour for a Store during creation.
type StoreOption func(*Store)

// WithModule namespaces a store under the provided module.
func WithModule(module string) StoreOption { return func(s *Store) { s.module = module } }

// WithPersistence enables localStorage persistence for the store.
func WithPersistence() StoreOption { return func(s *Store) { s.persist = true } }

// WithDevTools enables logging of state mutations for development.
func WithDevTools() StoreOption { return func(s *Store) { s.devTools = true } }

type Store struct {
	module     string
	name       string
	state      map[string]interface{}
	listeners  map[string]map[int]func(interface{})
	listenerID int
	computeds  map[string]*Computed
	watchers   []*Watcher
	persist    bool
	devTools   bool
}

type StoreManager struct {
	modules map[string]map[string]*Store
}

var GlobalStoreManager = &StoreManager{
	modules: make(map[string]map[string]*Store),
}

// NewStore creates a new store with the given name and optional configuration.
// By default stores are registered under the "default" module.
func NewStore(name string, opts ...StoreOption) *Store {
	store := &Store{
		module:    "default",
		name:      name,
		state:     make(map[string]interface{}),
		listeners: make(map[string]map[int]func(interface{})),
		computeds: make(map[string]*Computed),
	}
	for _, opt := range opts {
		opt(store)
	}

	GlobalStoreManager.RegisterStore(store.module, name, store)

	if store.persist {
		if state := loadPersistedState(store.storageKey()); state != nil {
			store.state = state
		}
	}

	return store
}

func (sm *StoreManager) RegisterStore(module, name string, store *Store) {
	if sm.modules[module] == nil {
		sm.modules[module] = make(map[string]*Store)
	}
	sm.modules[module][name] = store
}

func (sm *StoreManager) GetStore(module, name string) *Store {
	if stores, ok := sm.modules[module]; ok {
		return stores[name]
	}
	return nil
}

// Snapshot returns a deep copy of all registered stores and their states.
func (sm *StoreManager) Snapshot() map[string]map[string]map[string]interface{} {
	snap := make(map[string]map[string]map[string]interface{})
	for module, stores := range sm.modules {
		snap[module] = make(map[string]map[string]interface{})
		for name, store := range stores {
			stateCopy := make(map[string]interface{})
			for k, v := range store.state {
				stateCopy[k] = v
			}
			snap[module][name] = stateCopy
		}
	}
	return snap
}

func (s *Store) storageKey() string { return s.module + ":" + s.name }

func (s *Store) Set(key string, value interface{}) {
	s.state[key] = value
	if s.devTools {
		fmt.Printf("%s/%s -> %s: %v\n", s.module, s.name, key, value)
	}
	if StoreHook != nil {
		StoreHook(s.module, s.name, key, value)
	}
	if listeners, exists := s.listeners[key]; exists {
		for _, listener := range listeners {
			listener(value)
		}
	}
	s.evaluateDependents(key)
	if s.persist {
		saveState(s.storageKey(), s.state)
	}
}

func (s *Store) Get(key string) interface{} {
	if s.devTools {
		fmt.Printf("Getting %s from %s/%s\n", key, s.module, s.name)
	}
	return s.state[key]
}

func (s *Store) OnChange(key string, listener func(interface{})) func() {
	if s.listeners[key] == nil {
		s.listeners[key] = make(map[int]func(interface{}))
	}
	s.listenerID++
	id := s.listenerID
	s.listeners[key][id] = listener

	fmt.Println("------")
	for moduleName, stores := range GlobalStoreManager.modules {
		for storeName, store := range stores {
			fmt.Printf("Store: %s/%s\n", moduleName, storeName)
			for key, value := range store.state {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}
	fmt.Println("------")

	return func() {
		delete(s.listeners[key], id)
	}
}

// RegisterComputed registers a computed value on the store. The computed value
// is evaluated immediately and whenever one of its dependencies changes.
func (s *Store) RegisterComputed(c *Computed) {
	s.computeds[c.Key()] = c
	val := c.Evaluate(s.state)
	s.state[c.Key()] = val
	c.lastDeps = snapshotDeps(s.state, c.Deps())
}

// RegisterWatcher registers a watcher that triggers when any of its
// dependencies change. If the dependency list is empty the watcher is triggered
// on every state update.
func (s *Store) RegisterWatcher(w *Watcher) {
	s.watchers = append(s.watchers, w)
}

// evaluateDependents re-evaluates computed values and triggers watchers for a
// given key.
func (s *Store) evaluateDependents(key string) {
	for _, c := range s.computeds {
		if contains(c.Deps(), key) {
			current := snapshotDeps(s.state, c.Deps())
			if c.lastDeps == nil || !reflect.DeepEqual(current, c.lastDeps) {
				val := c.Evaluate(s.state)
				s.state[c.Key()] = val
				c.lastDeps = current
				if listeners, exists := s.listeners[c.Key()]; exists {
					for _, listener := range listeners {
						listener(val)
					}
				}
				// propagate to computeds/watchers depending on this key
				s.evaluateDependents(c.Key())
			}
		}
	}
	for _, w := range s.watchers {
		deps := w.Deps()
		if len(deps) == 0 || contains(deps, key) {
			w.Run(s.state)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func snapshotDeps(state map[string]interface{}, deps []string) map[string]interface{} {
	snap := make(map[string]interface{}, len(deps))
	for _, d := range deps {
		snap[d] = state[d]
	}
	return snap
}
