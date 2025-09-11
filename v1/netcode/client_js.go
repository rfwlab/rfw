//go:build js && wasm

package netcode

import hostclient "github.com/rfwlab/rfw/v1/hostclient"

// NewClient creates a netcode client bound to the given component name.
func NewClient[T any](name string, decode func(map[string]any) T, interp func(T, T, float64) T) *Client[T] {
	return newClient(name, decode, interp, hostclient.Send, hostclient.RegisterHandler)
}
