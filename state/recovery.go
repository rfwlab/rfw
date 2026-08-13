package state

import (
	"log"
	"runtime/debug"
)

// OnCallbackPanic receives panics recovered at state callback boundaries.
// Core installs the browser error pipeline here; packages that use state on
// their own still get a log entry instead of a process-ending panic.
var OnCallbackPanic func(recovered any, context string, stack []byte)

func reportCallbackPanic(recovered any, context string, stack []byte) {
	if OnCallbackPanic == nil {
		log.Printf("[rfw] recovered panic in %s: %v\n%s", context, recovered, stack)
		return
	}
	defer func() {
		if reporterPanic := recover(); reporterPanic != nil {
			log.Printf("[rfw] callback panic reporter failed: %v", reporterPanic)
			log.Printf("[rfw] recovered panic in %s: %v\n%s", context, recovered, stack)
		}
	}()
	OnCallbackPanic(recovered, context, stack)
}

func runCallback(context string, fn func()) (ok bool) {
	ok = true
	defer func() {
		if recovered := recover(); recovered != nil {
			ok = false
			reportCallbackPanic(recovered, context, debug.Stack())
		}
	}()
	fn()
	return ok
}

func runValueCallback[T any](context string, fn func() T) (value T, ok bool) {
	value, recovered, stack := captureValuePanic(fn)
	if recovered != nil {
		reportCallbackPanic(recovered, context, stack)
		return value, false
	}
	return value, true
}

func captureValuePanic[T any](fn func() T) (value T, recovered any, stack []byte) {
	defer func() {
		if value := recover(); value != nil {
			recovered = value
			stack = debug.Stack()
		}
	}()
	value = fn()
	return value, nil, nil
}
