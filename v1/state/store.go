package state

import "fmt"

type Store struct {
	name       string
	state      map[string]interface{}
	listeners  map[string]map[int]func(interface{})
	listenerID int
	computeds  map[string]*Computed
	watchers   []*Watcher
}

type StoreManager struct {
	stores map[string]*Store
}

var GlobalStoreManager = &StoreManager{
	stores: make(map[string]*Store),
}

func NewStore(name string) *Store {
	store := &Store{
		name:      name,
		state:     make(map[string]interface{}),
		listeners: make(map[string]map[int]func(interface{})),
		computeds: make(map[string]*Computed),
	}
	GlobalStoreManager.RegisterStore(name, store)
	return store
}

func (sm *StoreManager) RegisterStore(name string, store *Store) {
	sm.stores[name] = store
}

func (sm *StoreManager) GetStore(name string) *Store {
	return sm.stores[name]
}

func (s *Store) Set(key string, value interface{}) {
	s.state[key] = value
	if listeners, exists := s.listeners[key]; exists {
		for _, listener := range listeners {
			listener(value)
		}
	}
	s.evaluateDependents(key)
}

func (s *Store) Get(key string) interface{} {
	fmt.Printf("Getting key %s from store %v\n", key, s.name)
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
	for storeName, store := range GlobalStoreManager.stores {
		fmt.Printf("Store: %s\n", storeName)

		for key, value := range store.state {
			fmt.Printf("  %s: %v\n", key, value)
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
			val := c.Evaluate(s.state)
			s.state[c.Key()] = val
			if listeners, exists := s.listeners[c.Key()]; exists {
				for _, listener := range listeners {
					listener(val)
				}
			}
			// propagate to computeds/watchers depending on this key
			s.evaluateDependents(c.Key())
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
