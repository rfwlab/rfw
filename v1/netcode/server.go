package netcode

import (
	"sync"

	"github.com/rfwlab/rfw/v1/host"
)

// Server applies commands and broadcasts state snapshots.
type Server[T any] struct {
	name  string
	state T
	apply func(*T, any)
	mu    sync.Mutex
}

// NewServer creates a Server for the given component name and initial state.
func NewServer[T any](name string, initial T, apply func(*T, any)) *Server[T] {
	return &Server[T]{name: name, state: initial, apply: apply}
}

// HostComponent returns the handler to register with host.Register.
func (s *Server[T]) HostComponent() *host.HostComponent {
	return host.NewHostComponent(s.name, func(payload map[string]any) any {
		if cmds, ok := payload["commands"].([]any); ok {
			s.mu.Lock()
			for _, cmd := range cmds {
				s.apply(&s.state, cmd)
			}
			s.mu.Unlock()
		}
		return nil
	})
}

// Snapshot returns the current state.
func (s *Server[T]) Snapshot() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Broadcast pushes the current state with tick to all clients.
func (s *Server[T]) Broadcast(tick int64) {
	s.mu.Lock()
	state := s.state
	s.mu.Unlock()
	host.Broadcast(s.name, map[string]any{"tick": tick, "state": state})
}
