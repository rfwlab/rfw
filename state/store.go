package state

import (
	"log"
	"reflect"
	"strings"
	"sync"
)

func valEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	case int64:
		if bv, ok := b.(int64); ok {
			return av == bv
		}
	case float32:
		if bv, ok := b.(float32); ok {
			return av == bv
		}
	}
	return reflect.DeepEqual(a, b)
}

func depsChanged(current, last map[string]any) bool {
	if len(current) != len(last) {
		return true
	}
	for k, v := range current {
		if lv, ok := last[k]; !ok || !valEqual(v, lv) {
			return true
		}
	}
	return false
}

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

// WithHistory enables mutation history with the provided limit.
// The limit controls how many past mutations are retained for undo/redo.
func WithHistory(limit int) StoreOption {
	return func(s *Store) {
		if limit > 0 {
			s.historyLimit = limit
		}
	}
}

// Store holds keyed state with listeners, computed values, watchers and
// optional history. All methods are safe for concurrent use: internal state is
// mutex-protected, and listeners/watchers are invoked outside the lock (on the
// goroutine that called Set), so they may call back into the store. Computed
// functions run under the lock and must only read the state map they receive,
// never call store methods.
type Store struct {
	mu         sync.RWMutex
	module     string
	name       string
	state      map[string]any
	listeners  map[string]map[int]func(any)
	listenerID int
	computeds  map[string]*Computed
	watchers   []*Watcher
	persist    bool
	devTools   bool

	history      []*mutation
	future       []*mutation
	historyLimit int
}

type mutation struct {
	key      string
	previous any
	next     any
}

type StoreManager struct {
	mu      sync.RWMutex
	modules map[string]map[string]*Store
}

var GlobalStoreManager = &StoreManager{
	modules: make(map[string]map[string]*Store),
}

// NewStoreManager creates a standalone manager for isolating store instances.
func NewStoreManager() *StoreManager {
	return &StoreManager{modules: make(map[string]map[string]*Store)}
}

// NewStore creates a new store with the given name and optional configuration.
// By default stores are registered under the "default" module.
func NewStore(name string, opts ...StoreOption) *Store {
	return GlobalStoreManager.NewStore(name, opts...)
}

func (sm *StoreManager) NewStore(name string, opts ...StoreOption) *Store {
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

	sm.RegisterStore(store.module, name, store)

	if store.persist {
		if state := loadPersistedState(store.storageKey()); state != nil {
			store.state = state
		}
	}

	return store
}

func (sm *StoreManager) RegisterStore(module, name string, store *Store) {

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.modules[module] == nil {
		sm.modules[module] = make(map[string]*Store)
	}
	sm.modules[module][name] = store
}

func (sm *StoreManager) GetStore(module, name string) *Store {

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if stores, ok := sm.modules[module]; ok {
		return stores[name]
	}
	return nil
}

// UnregisterStore removes the store identified by module and name.
// If the store or module does not exist, it is a no-op.
func (sm *StoreManager) UnregisterStore(module, name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if stores, ok := sm.modules[module]; ok {
		delete(stores, name)
		if len(stores) == 0 {
			delete(sm.modules, module)
		}
	}
}

// Snapshot returns a deep copy of all registered stores and their states.
func (sm *StoreManager) Snapshot() map[string]map[string]map[string]any {
	snap := make(map[string]map[string]map[string]any)

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for module, stores := range sm.modules {
		snap[module] = make(map[string]map[string]any)
		for name, store := range stores {
			snap[module][name] = store.Snapshot()
		}
	}
	return snap
}

// Module reports the module namespace of the store.
func (s *Store) Module() string { return s.module }

// Name returns the store name within its module namespace.
func (s *Store) Name() string { return s.name }

// Snapshot copies the current state of the store.
func (s *Store) Snapshot() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap := make(map[string]any, len(s.state))
	for k, v := range s.state {
		snap[k] = v
	}
	return snap
}

func (s *Store) storageKey() string { return s.module + ":" + s.name }

func (s *Store) Set(key string, value any) {
	s.set(key, value, true)
}

// set applies a mutation under the lock, then fires listeners, watchers and
// persistence outside it so callbacks can safely call back into the store.
func (s *Store) set(key string, value any, recordHistory bool) {
	s.mu.Lock()
	old := s.state[key]
	s.state[key] = value
	if recordHistory && s.historyLimit > 0 {
		s.history = append(s.history, &mutation{key: key, previous: old, next: value})
		if len(s.history) > s.historyLimit {
			s.history = s.history[len(s.history)-s.historyLimit:]
		}
		s.future = nil
	}
	notifs := s.listenerNotifsLocked(key, value)
	notifs = append(notifs, s.evaluateDependentsLocked(key)...)
	var persisted map[string]any
	if s.persist {
		persisted = make(map[string]any, len(s.state))
		for k, v := range s.state {
			persisted[k] = v
		}
	}
	s.mu.Unlock()

	if s.devTools {
		logger.Debug("%s/%s -> %s: %v", s.module, s.name, key, value)
	}
	if StoreHook != nil {
		StoreHook(s.module, s.name, key, value)
	}
	for _, fn := range notifs {
		fn()
	}
	if persisted != nil {
		saveState(s.storageKey(), persisted)
	}
}

// listenerNotifsLocked snapshots the listeners registered for key as
// notification closures. Callers must hold s.mu.
func (s *Store) listenerNotifsLocked(key string, value any) []func() {
	listeners, exists := s.listeners[key]
	if !exists {
		return nil
	}
	notifs := make([]func(), 0, len(listeners))
	for _, listener := range listeners {
		l := listener
		notifs = append(notifs, func() { l(value) })
	}
	return notifs
}

func (s *Store) Get(key string) any {
	if s.devTools {
		logger.Debug("Getting %s from %s/%s", key, s.module, s.name)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state[key]
}

// Undo reverts the last mutation recorded in the store's history.
func (s *Store) Undo() {
	s.mu.Lock()
	if len(s.history) == 0 {
		s.mu.Unlock()
		return
	}
	m := s.history[len(s.history)-1]
	s.history = s.history[:len(s.history)-1]
	s.future = append(s.future, m)
	s.mu.Unlock()
	s.set(m.key, m.previous, false)
}

// Redo reapplies the last mutation that was undone.
func (s *Store) Redo() {
	s.mu.Lock()
	if len(s.future) == 0 {
		s.mu.Unlock()
		return
	}
	m := s.future[len(s.future)-1]
	s.future = s.future[:len(s.future)-1]
	s.history = append(s.history, m)
	if s.historyLimit > 0 && len(s.history) > s.historyLimit {
		s.history = s.history[len(s.history)-s.historyLimit:]
	}
	s.mu.Unlock()
	s.set(m.key, m.next, false)
}

func (s *Store) OnChange(key string, listener func(any)) func() {
	s.mu.Lock()
	if s.listeners[key] == nil {
		s.listeners[key] = make(map[int]func(any))
	}
	s.listenerID++
	id := s.listenerID
	s.listeners[key][id] = listener
	s.mu.Unlock()

	if s.devTools {
		logger.Debug("[rfw] store %s.%s: listener registered for key %s", s.module, s.name, key)
	}

	return func() {
		s.mu.Lock()
		delete(s.listeners[key], id)
		s.mu.Unlock()
	}
}

// RegisterComputed registers a computed value on the store. The computed value
// is evaluated immediately and whenever one of its dependencies changes. The
// compute function runs under the store lock: it must only read the state map
// it receives and never call store methods.
func (s *Store) RegisterComputed(c *Computed) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.computeds[c.Key()] = c
	val := c.Evaluate(s.state)
	s.state[c.Key()] = val
	c.lastDeps = snapshotDeps(s.state, c.Deps())
}

// Map registers a computed value derived from a single dependency using a
// strongly typed mapping function. The mapping function receives the current
// value of the dependency and its result is stored under the provided key. If
// the dependency cannot be asserted to the expected type, the zero value of the
// return type is used instead.
func Map[T, R any](s *Store, key, dep string, compute func(T) R) {
	c := NewComputed(key, []string{dep}, func(m map[string]any) any {
		if v, ok := m[dep].(T); ok {
			return compute(v)
		}
		var zero R
		return zero
	})
	s.RegisterComputed(c)
}

// Map2 registers a computed value derived from two dependencies. The mapping
// function receives the current values of both dependencies and its result is
// stored under the provided key. If any dependency fails type assertion the
// zero value of the return type is used.
func Map2[A, B, R any](s *Store, key, depA, depB string, compute func(A, B) R) {
	c := NewComputed(key, []string{depA, depB}, func(m map[string]any) any {
		a, okA := m[depA].(A)
		b, okB := m[depB].(B)
		if okA && okB {
			return compute(a, b)
		}
		var zero R
		return zero
	})
	s.RegisterComputed(c)
}

// RegisterWatcher registers a watcher that triggers when any of its
// dependencies change. If the dependency list is empty the watcher is triggered
// on every state update. It returns a function that removes the watcher.
func (s *Store) RegisterWatcher(w *Watcher) func() {
	s.mu.Lock()
	s.watchers = append(s.watchers, w)
	var snap map[string]any
	if w.immediate {
		snap = s.snapshotLocked()
	}
	s.mu.Unlock()
	if w.immediate {
		w.Run(snap)
	}

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, watcher := range s.watchers {
			if watcher == w {
				s.watchers = append(s.watchers[:i], s.watchers[i+1:]...)
				break
			}
		}
	}
}

// snapshotLocked copies the state map. Callers must hold s.mu.
func (s *Store) snapshotLocked() map[string]any {
	snap := make(map[string]any, len(s.state))
	for k, v := range s.state {
		snap[k] = v
	}
	return snap
}

// evaluateDependentsLocked re-evaluates computed values for a given key and
// collects listener/watcher notifications as closures to run after the lock is
// released. Watchers receive a consistent snapshot of the state taken at
// evaluation time. Callers must hold s.mu.
func (s *Store) evaluateDependentsLocked(key string) []func() {
	var notifs []func()
	for _, c := range s.computeds {
		if contains(c.Deps(), key) {
			current := snapshotDeps(s.state, c.Deps())
			if c.lastDeps == nil || depsChanged(current, c.lastDeps) {
				val := c.Evaluate(s.state)
				s.state[c.Key()] = val
				c.lastDeps = current
				notifs = append(notifs, s.listenerNotifsLocked(c.Key(), val)...)
				// propagate to computeds/watchers depending on this key
				notifs = append(notifs, s.evaluateDependentsLocked(c.Key())...)
			}
		}
	}
	var watcherSnap map[string]any
	runWatcher := func(w *Watcher) {
		if watcherSnap == nil {
			watcherSnap = s.snapshotLocked()
		}
		snap := watcherSnap
		notifs = append(notifs, func() { w.Run(snap) })
	}
	for _, w := range s.watchers {
		deps := w.Deps()
		if len(deps) == 0 {
			runWatcher(w)
			continue
		}
		for _, dep := range deps {
			if w.deep {
				if pathMatches(key, dep) {
					runWatcher(w)
					break
				}
			} else {
				if key == dep {
					runWatcher(w)
					break
				}
			}
		}
	}
	return notifs
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

func (sm *StoreManager) DumpState() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for mod, stores := range sm.modules {
		for name, st := range stores {
			logger.Debug("[rfw] DumpState %s/%s: %v", mod, name, st.Snapshot())
		}
	}
}
