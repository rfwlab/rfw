//go:build js && wasm

package core

// HostRegistrar registers a component with the SSC runtime and returns an
// idempotent cleanup that releases the registration. A nil cleanup means the
// registration owns nothing to release.
type HostRegistrar func(id, name string, vars []string) func()

// hostRegisterComponent registers a component with the SSC runtime. It stays
// nil until the hostclient package is linked, so a client-only build never
// reaches the websocket and net/http stacks through core.
var hostRegisterComponent HostRegistrar

// SetHostRegister installs the SSC component registration hook. The hostclient
// package calls it from its own init; applications never call it directly.
// Registrations made through it are never released, so prefer
// SetHostRegistrar when the runtime can unbind a component.
func SetHostRegister(fn func(id, name string, vars []string)) {
	if fn == nil {
		hostRegisterComponent = nil
		return
	}
	hostRegisterComponent = func(id, name string, vars []string) func() {
		fn(id, name, vars)
		return nil
	}
}

// SetHostRegistrar installs a cleanup-capable SSC registration hook, so a
// component that unmounts can release its binding, pending initialization and
// server subscription.
func SetHostRegistrar(fn HostRegistrar) {
	hostRegisterComponent = fn
}
