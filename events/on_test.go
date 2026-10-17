//go:build js && wasm

package events

import (
	"testing"

	js "github.com/rfwlab/rfw/v2/js"
)

// A released callback can hide a leaked native listener: inspect console.error
// as well as handler calls after removal.
func checkOnStop(t *testing.T, opts ...any) {
	t.Helper()
	target := js.Document().Call("createElement", "div")
	calls := 0
	stop := On("ping", target, func(js.Value) { calls++ }, opts...)
	dispatch(target, "ping")
	if calls != 1 {
		t.Fatalf("handler ran %d times", calls)
	}
	stop()
	console := js.Global().Get("console")
	original := console.Get("error")
	errors := 0
	report := js.SafeFuncOf(func(js.Value, []js.Value) any { errors++; return nil })
	console.Set("error", report)
	defer func() { console.Set("error", original); report.Release() }()
	dispatch(target, "ping")
	if calls != 1 || errors != 0 {
		t.Fatalf("after stop: calls=%d console errors=%d", calls, errors)
	}
}

func TestOnStopDefaultListener(t *testing.T) { checkOnStop(t) }

func TestOnStopCapturingListener(t *testing.T) { checkOnStop(t, true) }

func TestOnStopBubblingListener(t *testing.T) { checkOnStop(t, false) }

func TestOnStopCapturingOptions(t *testing.T) {
	checkOnStop(t, map[string]any{"capture": true, "passive": true})
}

func TestOnStopBubblingOptions(t *testing.T) {
	checkOnStop(t, map[string]any{"capture": false})
}

func TestOnStopKeepsRegisteredCaptureAfterOptionsChange(t *testing.T) {
	target := js.Document().Call("createElement", "div")
	options := js.ValueOf(map[string]any{"capture": true})
	calls := 0
	stop := On("ping", target, func(js.Value) { calls++ }, options)
	options.Set("capture", false)
	stop()
	console := js.Global().Get("console")
	original := console.Get("error")
	errors := 0
	report := js.SafeFuncOf(func(js.Value, []js.Value) any { errors++; return nil })
	console.Set("error", report)
	defer func() { console.Set("error", original); report.Release() }()
	dispatch(target, "ping")
	if calls != 0 || errors != 0 {
		t.Fatalf("after changed options: calls=%d console errors=%d", calls, errors)
	}
}
