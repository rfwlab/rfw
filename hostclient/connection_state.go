package hostclient

import "github.com/rfwlab/rfw/v2/state"

// ConnectionState describes the SSC transport state.
type ConnectionState string

const (
	ConnectionDisconnected ConnectionState = "disconnected"
	ConnectionConnecting   ConnectionState = "connecting"
	ConnectionConnected    ConnectionState = "connected"
	ConnectionDesynced     ConnectionState = "desynced"
)

var connectionState = state.NewSignal(ConnectionDisconnected)

// ConnectionStateSignal returns the reactive SSC connection state.
func ConnectionStateSignal() *state.Signal[ConnectionState] {
	return connectionState
}
