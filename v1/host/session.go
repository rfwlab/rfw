package host

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/rfwlab/rfw/v1/state"
)

// Session represents per-connection state for a WebSocket client.
// It exposes an isolated StoreManager and a context bag for arbitrary data.
type Session struct {
	id     string
	stores *state.StoreManager

	ctxMu sync.RWMutex
	ctx   map[string]any
}

func newSession(id string) *Session {
	return &Session{
		id:     id,
		stores: state.NewStoreManager(),
		ctx:    make(map[string]any),
	}
}

func (s *Session) ID() string { return s.id }

func (s *Session) StoreManager() *state.StoreManager { return s.stores }

// ContextGet retrieves a value from the session context.
func (s *Session) ContextGet(key string) (any, bool) {
	s.ctxMu.RLock()
	defer s.ctxMu.RUnlock()
	v, ok := s.ctx[key]
	return v, ok
}

// ContextSet stores a value in the session context.
func (s *Session) ContextSet(key string, value any) {
	s.ctxMu.Lock()
	s.ctx[key] = value
	s.ctxMu.Unlock()
}

// ContextDelete removes a value from the session context.
func (s *Session) ContextDelete(key string) {
	s.ctxMu.Lock()
	delete(s.ctx, key)
	s.ctxMu.Unlock()
}

// Snapshot returns a copy of all stores registered in this session.
func (s *Session) Snapshot() map[string]map[string]map[string]any {
	return s.stores.Snapshot()
}

var (
	sessionMu sync.RWMutex
	sessions  = make(map[string]*Session)
)

func allocateSession() *Session {
	id := generateSessionID()
	session := newSession(id)
	sessionMu.Lock()
	sessions[id] = session
	sessionMu.Unlock()
	return session
}

func releaseSession(session *Session) {
	if session == nil {
		return
	}
	sessionMu.Lock()
	delete(sessions, session.id)
	sessionMu.Unlock()
}

// SessionByID retrieves a session for the given ID.
func SessionByID(id string) (*Session, bool) {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	s, ok := sessions[id]
	return s, ok
}

func generateSessionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
