//go:build js && wasm

package core

// hostRegisterComponent registers a component with the SSC runtime. It stays
// nil until the hostclient package is linked, so a client-only build never
// reaches the websocket and net/http stacks through core.
var hostRegisterComponent func(id, name string, vars []string)

// SetHostRegister installs the SSC component registration hook. The hostclient
// package calls it from its own init; applications never call it directly.
func SetHostRegister(fn func(id, name string, vars []string)) {
	hostRegisterComponent = fn
}
