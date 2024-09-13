package framework

type Store struct {
	state      map[string]interface{}
	listeners  map[string]map[int]func(interface{})
	listenerID int
}

var stores = make(map[string]*Store)

func NewStore(name string) *Store {
	store := &Store{
		state:     make(map[string]interface{}),
		listeners: make(map[string]map[int]func(interface{})),
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

func (s *Store) OnChange(key string, listener func(interface{})) func() {
	if s.listeners[key] == nil {
		s.listeners[key] = make(map[int]func(interface{}))
	}
	s.listenerID++
	id := s.listenerID
	s.listeners[key][id] = listener
	return func() {
		delete(s.listeners[key], id)
	}
}
