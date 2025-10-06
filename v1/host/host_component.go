package host

import "github.com/rfwlab/rfw/v1/state"

// Handler processes inbound payloads for a HostComponent and returns a
// response payload to send back to the wasm runtime. Returning nil results in
// no message being sent.
type Handler func(payload map[string]any) any

// HandlerWithSession processes inbound payloads with the associated Session.
type HandlerWithSession func(*Session, map[string]any) any

// HostComponent represents server-side logic backing an HTML component.
type HostComponent struct {
	name           string
	handler        Handler
	sessionHandler HandlerWithSession
	initSnapshot   func(*Session, map[string]any) *InitSnapshot
}

// InitSnapshot represents markup the host can send to force the client to repaint a fragment.
type InitSnapshot struct {
	HTML string   `json:"html"`
	Vars []string `json:"vars,omitempty"`
}

// NewHostComponent registers a handler for the given component name.
func NewHostComponent(name string, handler Handler) *HostComponent {
	hc := &HostComponent{name: name, handler: handler}
	if handler != nil {
		hc.sessionHandler = func(_ *Session, payload map[string]any) any {
			return handler(payload)
		}
	}
	return hc
}

// WithInitSnapshot registers a callback that produces an InitSnapshot when a resync is requested.
func (hc *HostComponent) WithInitSnapshot(fn func(*Session, map[string]any) *InitSnapshot) *HostComponent {
	hc.initSnapshot = fn
	return hc
}

func (hc *HostComponent) Name() string { return hc.name }

// Handle executes the component's handler.
func (hc *HostComponent) Handle(payload map[string]any) any {
	if hc.handler == nil {
		return nil
	}
	return hc.handler(payload)
}

// NewHostComponentWithSession registers a session-aware handler.
func NewHostComponentWithSession(name string, handler HandlerWithSession) *HostComponent {
	return &HostComponent{name: name, sessionHandler: handler}
}

// HandleWithSession executes the session-aware handler when available.
func (hc *HostComponent) HandleWithSession(session *Session, payload map[string]any) any {
	if payload != nil {
		if _, ok := payload["resync"]; ok && hc.initSnapshot != nil {
			if snap := hc.initSnapshot(session, payload); snap != nil {
				return snap
			}
		}
	}
	if hc.sessionHandler != nil {
		return hc.sessionHandler(session, payload)
	}
	if hc.handler != nil {
		return hc.handler(payload)
	}
	return nil
}

// SessionAware reports whether the component registered a session handler.
func (hc *HostComponent) SessionAware() bool { return hc.sessionHandler != nil }

// StoreManager returns the session-specific store manager when available.
// If session is nil a reference to the global manager is returned for
// backward compatibility with legacy handlers.
func (hc *HostComponent) StoreManager(session *Session) *state.StoreManager {
	if session != nil {
		return session.StoreManager()
	}
	return state.GlobalStoreManager
}

var registry = make(map[string]*HostComponent)

// Register adds a HostComponent to the global registry so incoming messages
// can be routed to it.
func Register(hc *HostComponent) { registry[hc.name] = hc }

// Get returns a registered HostComponent by name.
func Get(name string) (*HostComponent, bool) { hc, ok := registry[name]; return hc, ok }
