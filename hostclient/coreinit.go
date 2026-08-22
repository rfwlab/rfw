//go:build js && wasm

package hostclient

import "github.com/rfwlab/rfw/v2/core"

// Linking this package wires SSC component registration into core. Apps that
// never use SSC never import hostclient, so the websocket and net/http stacks
// stay out of their bundle.
func init() { core.SetHostRegister(RegisterComponent) }
