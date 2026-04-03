package server

import signalbus "github.com/mirkobrombin/go-signal/v2/pkg/bus"

// RebuildEvent is emitted when a file change triggers a rebuild.
type RebuildEvent struct {
	Path    string // triggering file path
	Success bool   // whether the rebuild succeeded
	Error   string // error message if failed
}

// RebuildBus is the event bus used to publish rebuild notifications.
// External code (tests, plugins) can subscribe to it.
var RebuildBus = signalbus.New(signalbus.WithStrategy(signalbus.BestEffort))
