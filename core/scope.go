package core

import (
	"context"
	"sync"
)

// Scope owns work that must stop with a component.
type Scope struct {
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	cleanups []func()
	closed   bool
}

// NewScope creates an open lifecycle scope.
func NewScope() *Scope {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scope{ctx: ctx, cancel: cancel}
}

// Context is cancelled when the scope closes.
func (s *Scope) Context() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ctx
}

// Defer registers cleanup work in last-in, first-out order.
func (s *Scope) Defer(fn func()) {
	if fn == nil {
		return
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		fn()
		return
	}
	s.cleanups = append(s.cleanups, fn)
	s.mu.Unlock()
}

// Go starts work with the scope context.
func (s *Scope) Go(fn func(context.Context)) {
	if fn == nil {
		return
	}
	go fn(s.Context())
}

// Close cancels the context and runs registered cleanup once.
func (s *Scope) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	cancel := s.cancel
	cleanups := s.cleanups
	s.cleanups = nil
	s.mu.Unlock()

	cancel()
	for i := len(cleanups) - 1; i >= 0; i-- {
		func(cleanup func()) {
			defer func() {
				if recovered := recover(); recovered != nil {
					reportScopeError(recovered)
				}
			}()
			cleanup()
		}(cleanups[i])
	}
}

// Closed reports whether Close has run.
func (s *Scope) Closed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}
