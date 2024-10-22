package framework

import "fmt"

type Store struct {
	name       string
	state      map[string]interface{}
	listeners  map[string]map[int]func(interface{})
	listenerID int
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
