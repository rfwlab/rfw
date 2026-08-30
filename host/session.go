package host

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/rfwlab/rfw/v2/state"
	"golang.org/x/net/websocket"
)

// Session represents per-connection state for a WebSocket client.
// It exposes an isolated StoreManager and a context bag for arbitrary data.
type Session struct {
	id          string
	resumeToken string
	stores      *state.StoreManager

	ctxMu sync.RWMutex
	ctx   map[string]any

	deliveryMu        sync.Mutex
	outboundMu        sync.Mutex
	connection        *websocket.Conn
	connectionManaged bool
	resumePending     bool
	attached          bool
	released          bool
	expires           time.Time
	expiryTimer       *time.Timer
	inboundSeq        uint64
	outboundSeq       uint64
	rateStart         time.Time
	rateCount         int
	replayLimit       int
	replay            []Outbound
}

type sessionOptions struct {
	resumeToken string
	replayLimit int
}

func newSession(id string, options ...sessionOptions) *Session {
	var config sessionOptions
	if len(options) > 0 {
		config = options[0]
	}
	return &Session{
		id:          id,
		resumeToken: config.resumeToken,
		stores:      state.NewStoreManager(),
		ctx:         make(map[string]any),
		attached:    true,
		replayLimit: config.replayLimit,
	}
}

// ID returns the session ID.
func (s *Session) ID() string { return s.id }

// ResumeToken returns the opaque token used to resume this session.
func (s *Session) ResumeToken() string { return s.resumeToken }

// StoreManager returns the session-local store registry.
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
	sessionMu      sync.RWMutex
	sessions       = make(map[string]*Session)
	sessionByToken = make(map[string]*Session)
)

// AllocateSession creates and registers a session.
func AllocateSession() *Session {
	session, _ := allocateSession(0, 0)
	return session
}

// AllocateResumableSession creates a session with ordered delivery history.
func AllocateResumableSession(replayLimit int) *Session {
	session, _ := allocateSession(replayLimit, 0)
	return session
}

func allocateSession(replayLimit, maxSessions int) (*Session, error) {
	id := generateSessionID()
	token := ""
	if replayLimit > 0 {
		token = generateSessionID() + generateSessionID()
	}
	session := newSession(id, sessionOptions{resumeToken: token, replayLimit: replayLimit})
	sessionMu.Lock()
	if maxSessions > 0 && len(sessions) >= maxSessions {
		sessionMu.Unlock()
		return nil, ErrSessionLimit
	}
	sessions[id] = session
	if token != "" {
		sessionByToken[token] = session
	}
	sessionMu.Unlock()
	return session, nil
}

// SuspendSession detaches a connection and retains resumable state for ttl.
func SuspendSession(session *Session, ttl time.Duration) {
	if session == nil {
		return
	}
	session.outboundMu.Lock()
	session.deliveryMu.Lock()
	if !session.attached {
		session.deliveryMu.Unlock()
		session.outboundMu.Unlock()
		return
	}
	session.connection = nil
	session.attached = false
	if ttl <= 0 || session.resumeToken == "" {
		session.deliveryMu.Unlock()
		session.outboundMu.Unlock()
		ReleaseSession(session)
		return
	}
	expires := time.Now().Add(ttl)
	session.expires = expires
	session.expiryTimer = time.AfterFunc(ttl, func() {
		releaseSession(session, expires)
	})
	session.deliveryMu.Unlock()
	session.outboundMu.Unlock()
}

// ResumeSession attaches a disconnected session by opaque token.
// The new socket must call ReplaySession or BindSessionConnection before sends.
func ResumeSession(token string) (*Session, bool) {
	session := sessionForToken(token)
	if session == nil {
		return nil, false
	}
	session.deliveryMu.Lock()
	defer session.deliveryMu.Unlock()
	if !session.resumableLocked() {
		return nil, false
	}
	session.markResumedLocked()
	return session, true
}

// resumeCandidate reports the session a token can currently resume without
// mutating it, so an authorization decision can be taken before the session is
// attached, its expiry cleared or its connection replaced.
func resumeCandidate(token string) (*Session, bool) {
	session := sessionForToken(token)
	if session == nil {
		return nil, false
	}
	session.deliveryMu.Lock()
	defer session.deliveryMu.Unlock()
	if !session.resumableLocked() {
		return nil, false
	}
	return session, true
}

// commitResume attaches a candidate that authorization approved. It fails
// closed when the token stopped mapping to that exact session or the session is
// no longer resumable, so a concurrent attempt cannot reattach a session other
// than the one that was authorized.
func commitResume(token string, candidate *Session) bool {
	if candidate == nil || sessionForToken(token) != candidate {
		return false
	}
	candidate.deliveryMu.Lock()
	defer candidate.deliveryMu.Unlock()
	if !candidate.resumableLocked() {
		return false
	}
	candidate.markResumedLocked()
	return true
}

func sessionForToken(token string) *Session {
	if token == "" {
		return nil
	}
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return sessionByToken[token]
}

// resumableLocked reports whether a detached session can still be reattached.
func (s *Session) resumableLocked() bool {
	return !s.released && !s.attached &&
		(s.expires.IsZero() || !time.Now().After(s.expires))
}

// markResumedLocked attaches the session and cancels its retention timer.
func (s *Session) markResumedLocked() {
	s.attached = true
	s.resumePending = true
	s.expires = time.Time{}
	if s.expiryTimer != nil {
		s.expiryTimer.Stop()
		s.expiryTimer = nil
	}
}

// ReleaseSession removes a session from the registry.
func ReleaseSession(session *Session) {
	releaseSession(session, time.Time{})
}

func releaseSession(session *Session, expectedExpiry time.Time) {
	if session == nil {
		return
	}
	session.outboundMu.Lock()
	session.deliveryMu.Lock()
	if session.released || (!expectedExpiry.IsZero() &&
		(session.attached || !session.expires.Equal(expectedExpiry))) {
		session.deliveryMu.Unlock()
		session.outboundMu.Unlock()
		return
	}
	session.released = true
	if session.expiryTimer != nil {
		session.expiryTimer.Stop()
		session.expiryTimer = nil
	}
	session.connection = nil
	session.connectionManaged = false
	session.resumePending = false
	session.attached = false
	session.deliveryMu.Unlock()
	session.outboundMu.Unlock()
	sessionMu.Lock()
	if sessions[session.id] == session {
		delete(sessions, session.id)
	}
	if session.resumeToken != "" && sessionByToken[session.resumeToken] == session {
		delete(sessionByToken, session.resumeToken)
	}
	sessionMu.Unlock()
}

// SessionByID retrieves a session for the given ID.
func SessionByID(id string) (*Session, bool) {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	s, ok := sessions[id]
	return s, ok
}

var (
	// ErrDuplicateMessage reports an already processed client sequence.
	ErrDuplicateMessage = errors.New("host: duplicate message")
	// ErrSequenceGap reports a missing client message.
	ErrSequenceGap = errors.New("host: message sequence gap")
	// ErrReplayUnavailable reports that acknowledged history is too old.
	ErrReplayUnavailable = errors.New("host: replay unavailable")
	// ErrSessionLimit reports that the retained-session limit was reached.
	ErrSessionLimit = errors.New("host: session limit reached")
)

// AcceptInbound validates and records an inbound sequence.
func (s *Session) AcceptInbound(sequence uint64) error {
	if s == nil || sequence == 0 {
		return nil
	}
	s.deliveryMu.Lock()
	defer s.deliveryMu.Unlock()
	if sequence <= s.inboundSeq {
		return ErrDuplicateMessage
	}
	if s.inboundSeq != 0 && sequence != s.inboundSeq+1 {
		return ErrSequenceGap
	}
	s.inboundSeq = sequence
	return nil
}

// AllowMessage enforces a fixed per-session message window.
func (s *Session) AllowMessage(limit int) bool {
	if s == nil || limit <= 0 {
		return true
	}
	s.deliveryMu.Lock()
	defer s.deliveryMu.Unlock()
	now := time.Now()
	if s.rateStart.IsZero() || now.Sub(s.rateStart) >= time.Minute {
		s.rateStart = now
		s.rateCount = 0
	}
	s.rateCount++
	return s.rateCount <= limit
}

// PrepareOutbound assigns delivery metadata and stores replay history.
func (s *Session) PrepareOutbound(out Outbound) Outbound {
	if s == nil {
		return out
	}
	s.deliveryMu.Lock()
	defer s.deliveryMu.Unlock()
	s.outboundSeq++
	out.Session = s.id
	out.Sequence = s.outboundSeq
	out.Ack = s.inboundSeq
	out.ResumeToken = s.resumeToken
	if s.replayLimit > 0 {
		s.replay = append(s.replay, out)
		if extra := len(s.replay) - s.replayLimit; extra > 0 {
			copy(s.replay, s.replay[extra:])
			s.replay = s.replay[:s.replayLimit]
		}
	}
	return out
}

// Acknowledge removes outbound messages confirmed by the client.
func (s *Session) Acknowledge(sequence uint64) {
	if s == nil || sequence == 0 {
		return
	}
	s.deliveryMu.Lock()
	defer s.deliveryMu.Unlock()
	remove := 0
	for remove < len(s.replay) && s.replay[remove].Sequence <= sequence {
		remove++
	}
	if remove > 0 {
		s.replay = append([]Outbound(nil), s.replay[remove:]...)
	}
}

// ReplayAfter returns retained outbound messages after sequence.
func (s *Session) ReplayAfter(sequence uint64) ([]Outbound, error) {
	if s == nil {
		return nil, nil
	}
	s.deliveryMu.Lock()
	defer s.deliveryMu.Unlock()
	if len(s.replay) == 0 {
		return nil, nil
	}
	if sequence+1 < s.replay[0].Sequence {
		return nil, ErrReplayUnavailable
	}
	index := 0
	for index < len(s.replay) && s.replay[index].Sequence <= sequence {
		index++
	}
	return append([]Outbound(nil), s.replay[index:]...), nil
}

func generateSessionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
