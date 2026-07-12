//go:build js && wasm

package core

import (
	"sync"

	"github.com/rfwlab/rfw/v2/dom"
)

// Runtime errors flow through a single pipeline: every capture point (Render,
// Mount, Unmount, navigation, effects, template loading, delegated event
// handlers and ErrorBoundary) recovers the panic and hands it to ReportError,
// which fans out to the registered sinks. The developer overlay is the
// default sink; applications add their own with OnError to forward errors to
// logging or telemetry.

// ErrorSink receives a recovered error together with a short human-readable
// context such as "Render: Home (ID: abc)".
type ErrorSink func(err any, context string)

var (
	errorMu    sync.Mutex
	errorSinks []ErrorSink
)

// OnError registers a sink invoked for every error reported by the runtime.
// It returns a function that removes the sink.
func OnError(fn ErrorSink) func() {
	if fn == nil {
		return func() {}
	}
	errorMu.Lock()
	errorSinks = append(errorSinks, fn)
	idx := len(errorSinks) - 1
	errorMu.Unlock()
	return func() {
		errorMu.Lock()
		if idx < len(errorSinks) {
			errorSinks[idx] = nil
		}
		errorMu.Unlock()
	}
}

// ReportError delivers err to every registered sink and to the developer
// overlay. All recovery paths in the framework funnel through here.
func ReportError(err any, context string) {
	errorMu.Lock()
	sinks := make([]ErrorSink, len(errorSinks))
	copy(sinks, errorSinks)
	errorMu.Unlock()
	for _, fn := range sinks {
		if fn != nil {
			fn(err, context)
		}
	}
	ShowErrorOverlay(err, context)
}

// Delegated event handlers recover panics inside the dom package; route them
// into the same pipeline as every other capture point.
func init() {
	dom.OnHandlerPanic = func(err any, name string) {
		ReportError(err, "Handler: "+name)
	}
}
