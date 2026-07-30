package host

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

// SSCLimits bounds WebSocket resource use and action execution.
type SSCLimits struct {
	MaxMessageBytes   int
	MaxConnections    int64
	MaxSessions       int
	MessagesPerMinute int
	HandlerTimeout    time.Duration
	ResumeTTL         time.Duration
	ReplayMessages    int
}

// DefaultSSCLimits returns the production defaults used by NewMux.
func DefaultSSCLimits() SSCLimits {
	return SSCLimits{
		MaxMessageBytes:   1 << 20,
		MaxConnections:    4096,
		MaxSessions:       8192,
		MessagesPerMinute: 600,
		HandlerTimeout:    15 * time.Second,
		ResumeTTL:         2 * time.Minute,
		ReplayMessages:    256,
	}
}

// MessageAuthorizer can reject any decoded SSC message.
type MessageAuthorizer func(context.Context, *Session, Inbound) error

// SessionInitializer copies authenticated request state into a new session.
type SessionInitializer func(*http.Request, *Session) error

// MuxOption configures the WebSocket endpoint created by NewMux.
type MuxOption func(*WSRuntime)

// WSRuntime holds the guards, limits, and connection count for one endpoint.
type WSRuntime struct {
	authFunc    func(*http.Request) bool
	origins     []string
	authorize   MessageAuthorizer
	initialize  SessionInitializer
	limits      SSCLimits
	connections atomic.Int64
}

// NewWSRuntime resolves MuxOptions into an endpoint runtime.
func NewWSRuntime(opts ...MuxOption) *WSRuntime {
	runtime := &WSRuntime{limits: DefaultSSCLimits()}
	for _, opt := range opts {
		opt(runtime)
	}
	return runtime
}

// WithAuthFunc registers a callback invoked before the WebSocket upgrade.
func WithAuthFunc(fn func(*http.Request) bool) MuxOption {
	return func(runtime *WSRuntime) { runtime.authFunc = fn }
}

// WithOriginAllowlist restricts upgrades to exact Origin matches.
func WithOriginAllowlist(origins ...string) MuxOption {
	return func(runtime *WSRuntime) {
		runtime.origins = append(runtime.origins, origins...)
	}
}

// WithSSCAuthorizer adds authorization after a message is decoded.
func WithSSCAuthorizer(authorize MessageAuthorizer) MuxOption {
	return func(runtime *WSRuntime) { runtime.authorize = authorize }
}

// WithSSCSessionInitializer initializes session identity from the upgrade request.
func WithSSCSessionInitializer(initialize SessionInitializer) MuxOption {
	return func(runtime *WSRuntime) { runtime.initialize = initialize }
}

// WithSSCLimits overrides non-zero SSC resource limits.
func WithSSCLimits(limits SSCLimits) MuxOption {
	return func(runtime *WSRuntime) {
		if limits.MaxMessageBytes > 0 {
			runtime.limits.MaxMessageBytes = limits.MaxMessageBytes
		}
		if limits.MaxConnections > 0 {
			runtime.limits.MaxConnections = limits.MaxConnections
		}
		if limits.MaxSessions > 0 {
			runtime.limits.MaxSessions = limits.MaxSessions
		}
		if limits.MessagesPerMinute > 0 {
			runtime.limits.MessagesPerMinute = limits.MessagesPerMinute
		}
		if limits.HandlerTimeout > 0 {
			runtime.limits.HandlerTimeout = limits.HandlerTimeout
		}
		if limits.ResumeTTL > 0 {
			runtime.limits.ResumeTTL = limits.ResumeTTL
		}
		if limits.ReplayMessages > 0 {
			runtime.limits.ReplayMessages = limits.ReplayMessages
		}
	}
}

// WithoutSSCResume releases sessions as soon as their connection closes.
func WithoutSSCResume() MuxOption {
	return func(runtime *WSRuntime) {
		runtime.limits.ResumeTTL = 0
		runtime.limits.ReplayMessages = 0
	}
}

// Guard applies origin and upgrade authentication checks.
func (runtime *WSRuntime) Guard(next http.Handler) http.Handler {
	if runtime == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(runtime.origins) > 0 {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, candidate := range runtime.origins {
				if origin == candidate {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}
		}
		if runtime.authFunc != nil && !runtime.authFunc(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GuardWS wraps a WebSocket handler using MuxOptions.
func GuardWS(next http.Handler, opts ...MuxOption) http.Handler {
	return NewWSRuntime(opts...).Guard(next)
}

// AcquireConnection reserves a connection slot.
func (runtime *WSRuntime) AcquireConnection() bool {
	if runtime == nil {
		return true
	}
	active := runtime.connections.Add(1)
	if runtime.limits.MaxConnections > 0 && active > runtime.limits.MaxConnections {
		runtime.connections.Add(-1)
		return false
	}
	return true
}

// ReleaseConnection frees a reserved connection slot.
func (runtime *WSRuntime) ReleaseConnection() {
	if runtime != nil {
		runtime.connections.Add(-1)
	}
}

// ConfigureConnection applies the frame-size limit.
func (runtime *WSRuntime) ConfigureConnection(ws *websocket.Conn) {
	if runtime != nil && runtime.limits.MaxMessageBytes > 0 {
		ws.MaxPayloadBytes = runtime.limits.MaxMessageBytes
	}
}

// NewSession allocates and initializes a resumable session.
func (runtime *WSRuntime) NewSession(request *http.Request) (*Session, error) {
	replayLimit := 0
	if runtime != nil {
		replayLimit = runtime.limits.ReplayMessages
	}
	maxSessions := 0
	if runtime != nil {
		maxSessions = runtime.limits.MaxSessions
	}
	session, err := allocateSession(replayLimit, maxSessions)
	if err != nil {
		return nil, err
	}
	if runtime != nil && runtime.initialize != nil {
		if err := runtime.initialize(request, session); err != nil {
			ReleaseSession(session)
			return nil, err
		}
	}
	return session, nil
}

// Authorize validates a decoded message.
func (runtime *WSRuntime) Authorize(ctx context.Context, session *Session, message Inbound) error {
	if runtime == nil || runtime.authorize == nil {
		return nil
	}
	return runtime.authorize(ctx, session, message)
}

// HandlerContext returns a context bounded by HandlerTimeout.
func (runtime *WSRuntime) HandlerContext(parent context.Context) (context.Context, context.CancelFunc) {
	if runtime == nil || runtime.limits.HandlerTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, runtime.limits.HandlerTimeout)
}

// ResumeTTL returns the configured detached-session lifetime.
func (runtime *WSRuntime) ResumeTTL() time.Duration {
	if runtime == nil {
		return 0
	}
	return runtime.limits.ResumeTTL
}

// MessagesPerMinute returns the configured per-session message limit.
func (runtime *WSRuntime) MessagesPerMinute() int {
	if runtime == nil {
		return 0
	}
	return runtime.limits.MessagesPerMinute
}
