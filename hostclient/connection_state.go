// Package hostclient connects browser components to the SSC host.
package hostclient

import "github.com/rfwlab/rfw/v2/state"

// ConnectionState describes the SSC transport state.
type ConnectionState string

const (
	// ConnectionDisconnected indicates that no SSC connection is active.
	ConnectionDisconnected ConnectionState = "disconnected"
	// ConnectionConnecting indicates that an SSC connection is opening.
	ConnectionConnecting ConnectionState = "connecting"
	// ConnectionConnected indicates that the SSC connection is ready.
	ConnectionConnected ConnectionState = "connected"
	// ConnectionDesynced indicates that SSC state must be synchronized.
	ConnectionDesynced ConnectionState = "desynced"
)

var connectionState = state.NewSignal(ConnectionDisconnected)

// ConnectionStateSignal returns the reactive SSC connection state.
func ConnectionStateSignal() *state.Signal[ConnectionState] {
	return connectionState
}
