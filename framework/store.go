package framework

type Store struct {
	state     map[string]interface{}
	listeners map[string][]func(interface{})
}

var stores = make(map[string]*Store)

func NewStore(name string) *Store {
	store := &Store{
		state:     make(map[string]interface{}),
		listeners: make(map[string][]func(interface{})),
	}
	stores[name] = store
	return store
}

func GetStore(name string) *Store {
	return stores[name]
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
	return s.state[key]
}

func (s *Store) OnChange(key string, listener func(interface{})) {
	s.listeners[key] = append(s.listeners[key], listener)
}
