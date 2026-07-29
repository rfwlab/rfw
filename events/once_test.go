//go:build js && wasm

package events

import (
	"testing"

	js "github.com/rfwlab/rfw/v2/js"
)

func dispatch(target js.Value, name string) {
	evt := js.CustomEvent().New(name)
	target.Call("dispatchEvent", evt)
}

func TestOnceFiresOnlyOnce(t *testing.T) {
	target := js.Document().Call("createElement", "div")
	calls := 0
	Once("ping", target, func(js.Value) { calls++ })

	dispatch(target, "ping")
	dispatch(target, "ping")

	if calls != 1 {
		t.Fatalf("handler ran %d times, want 1", calls)
	}
}

func TestOnceCancelBeforeFiring(t *testing.T) {
	target := js.Document().Call("createElement", "div")
	calls := 0
	cancel := Once("ping", target, func(js.Value) { calls++ })
	cancel()

	dispatch(target, "ping")

	if calls != 0 {
		t.Fatalf("handler ran %d times after cancel, want 0", calls)
	}
}

func TestOnceCancelAfterFiringIsNoop(t *testing.T) {
	target := js.Document().Call("createElement", "div")
	cancel := Once("ping", target, func(js.Value) {})
	dispatch(target, "ping")
	cancel()
	cancel()
}
