//go:build js && wasm

package events

import (
	"testing"

	js "github.com/rfwlab/rfw/v2/js"
)

// A panicking handler must not take the wasm instance down with it: the
// framework's own listeners carry the recover guard, so an application that
// uses events.On never has to reach for SafeFuncOf itself.
func TestOnRecoversHandlerPanic(t *testing.T) {
	prev := js.OnFuncPanic
	defer func() { js.OnFuncPanic = prev }()

	var got any
	js.OnFuncPanic = func(r any, _ []byte) { got = r }

	target := js.Document().Call("createElement", "div")
	stop := On("boom", target, func(js.Value) { panic("handler exploded") })
	defer stop()

	target.Call("dispatchEvent", js.CustomEvent().New("boom"))

	if got == nil {
		t.Fatal("panic escaped the listener")
	}
	if s, ok := got.(string); !ok || s != "handler exploded" {
		t.Fatalf("unexpected recovered value: %v", got)
	}
}

func TestOnceRecoversHandlerPanic(t *testing.T) {
	prev := js.OnFuncPanic
	defer func() { js.OnFuncPanic = prev }()

	var got any
	js.OnFuncPanic = func(r any, _ []byte) { got = r }

	target := js.Document().Call("createElement", "div")
	Once("boom", target, func(js.Value) { panic("once exploded") })
	target.Call("dispatchEvent", js.CustomEvent().New("boom"))

	if got == nil {
		t.Fatal("panic escaped the one-shot listener")
	}
}
