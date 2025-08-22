package host

// Handler processes inbound payloads for a HostComponent and returns a
// response payload to send back to the wasm runtime. Returning nil results in
// no message being sent.
type Handler func(payload map[string]any) any

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
func (hc *HostComponent) Handle(payload map[string]any) any {
	if hc.handler != nil {
		return hc.handler(payload)
	}
	return nil
}

var registry = make(map[string]*HostComponent)

// Register adds a HostComponent to the global registry so incoming messages
// can be routed to it.
func Register(hc *HostComponent) { registry[hc.name] = hc }

// Get returns a registered HostComponent by name.
func Get(name string) (*HostComponent, bool) { hc, ok := registry[name]; return hc, ok }
