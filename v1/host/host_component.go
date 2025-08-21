package host

import (
	"encoding/json"

	"golang.org/x/net/websocket"
)

// Context wraps a WebSocket connection and the component name so that
// host components can easily notify the connected wasm runtime.
// Messages are serialized as {"component": "name", "payload": any}.
type Context struct {
	conn      *websocket.Conn
	component string
}

// NewContext creates a communication context for the given WebSocket
// connection and component name.
func NewContext(conn *websocket.Conn, component string) *Context {
	return &Context{conn: conn, component: component}
}

// Notify sends a payload back to the wasm client for this component.
func (c *Context) Notify(payload any) error {
	msg := struct {
		Component string `json:"component"`
		Payload   any    `json:"payload"`
	}{Component: c.component, Payload: payload}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return websocket.Message.Send(c.conn, b)
}

// Handler processes inbound payloads for a HostComponent.
type Handler func(ctx *Context, payload json.RawMessage)

// HostComponent represents server-side logic backing an HTML component.
type HostComponent struct {
	name    string
	handler Handler
}

// NewHostComponent registers a handler for the given component name.
func NewHostComponent(name string, handler Handler) *HostComponent {
	return &HostComponent{name: name, handler: handler}
}

func (hc *HostComponent) Name() string { return hc.name }

// Handle executes the component's handler.
func (hc *HostComponent) Handle(ctx *Context, payload json.RawMessage) {
	if hc.handler != nil {
		hc.handler(ctx, payload)
	}
}

var registry = make(map[string]*HostComponent)

// Register adds a HostComponent to the global registry so incoming messages
// can be routed to it.
func Register(hc *HostComponent) { registry[hc.name] = hc }

// Get returns a registered HostComponent by name.
func Get(name string) (*HostComponent, bool) { hc, ok := registry[name]; return hc, ok }
