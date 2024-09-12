package framework

type Store struct {
	state     map[string]interface{}
	listeners map[string][]func(interface{})
}

var globalStore *Store

func NewStore() *Store {
	return &Store{
		state:     make(map[string]interface{}),
		listeners: make(map[string][]func(interface{})),
	}
}

func InitGlobalStore() {
	globalStore = NewStore()
}

func GetGlobalStore() *Store {
	return globalStore
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
