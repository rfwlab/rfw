package state

import (
	"log"
	"reflect"
	"strings"
)

type Logger interface {
	Debug(format string, args ...any)
}

type defaultLogger struct{}

func (defaultLogger) Debug(format string, args ...any) { log.Printf(format, args...) }

var logger Logger = defaultLogger{}

func SetLogger(l Logger) { logger = l }

// StoreHook, if non-nil, is invoked on every mutation allowing external
// observers (e.g. plugins) to react to state changes without creating an
// import cycle with core.
var StoreHook func(module, store, key string, value any)

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
	state      map[string]any
	listeners  map[string]map[int]func(any)
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
		state:     make(map[string]any),
		listeners: make(map[string]map[int]func(any)),
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
func (sm *StoreManager) Snapshot() map[string]map[string]map[string]any {
	snap := make(map[string]map[string]map[string]any)
	for module, stores := range sm.modules {
		snap[module] = make(map[string]map[string]any)
		for name, store := range stores {
			stateCopy := make(map[string]any)
			for k, v := range store.state {
				stateCopy[k] = v
			}
			snap[module][name] = stateCopy
		}
	}
	return snap
}

func (s *Store) storageKey() string { return s.module + ":" + s.name }

func (s *Store) Set(key string, value any) {
	s.state[key] = value
	if s.devTools {
		logger.Debug("%s/%s -> %s: %v", s.module, s.name, key, value)
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

func (s *Store) Get(key string) any {
	if s.devTools {
		logger.Debug("Getting %s from %s/%s", key, s.module, s.name)
	}
	return s.state[key]
}

func (s *Store) OnChange(key string, listener func(any)) func() {
	if s.listeners[key] == nil {
		s.listeners[key] = make(map[int]func(any))
	}
	s.listenerID++
	id := s.listenerID
	s.listeners[key][id] = listener

	logger.Debug("------")
	for moduleName, stores := range GlobalStoreManager.modules {
		for storeName, store := range stores {
			logger.Debug("Store: %s/%s", moduleName, storeName)
			for key, value := range store.state {
				logger.Debug("  %s: %v", key, value)
			}
		}
	}
	logger.Debug("------")

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
// on every state update. It returns a function that removes the watcher.
func (s *Store) RegisterWatcher(w *Watcher) func() {
	s.watchers = append(s.watchers, w)
	if w.immediate {
		w.Run(s.state)
	}

	return func() {
		for i, watcher := range s.watchers {
			if watcher == w {
				s.watchers = append(s.watchers[:i], s.watchers[i+1:]...)
				break
			}
		}
	}
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
		if len(deps) == 0 {
			w.Run(s.state)
			continue
		}
		for _, dep := range deps {
			if w.deep {
				if pathMatches(key, dep) {
					w.Run(s.state)
					break
				}
			} else {
				if key == dep {
					w.Run(s.state)
					break
				}
			}
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

func pathMatches(key, dep string) bool {
	if key == dep {
		return true
	}
	if strings.HasPrefix(key, dep+".") {
		return true
	}
	if strings.HasPrefix(dep, key+".") {
		return true
	}
	return false
}

func snapshotDeps(state map[string]any, deps []string) map[string]any {
	snap := make(map[string]any, len(deps))
	for _, d := range deps {
		snap[d] = state[d]
	}
	return snap
}
