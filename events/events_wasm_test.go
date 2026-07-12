//go:build js && wasm

package events

import (
	"testing"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

// Unsubscribe must remove exactly the handler it was returned for; the old
// implementation compared local variable addresses and never removed anything.
func TestOnAppUnsubscribe(t *testing.T) {
	const ev Event = "test:unsub"
	var first, second int
	off1 := OnApp(ev, func(any) { first++ })
	off2 := OnApp(ev, func(any) { second++ })

	EmitApp(ev, nil)
	off1()
	EmitApp(ev, nil)
	off2()
	EmitApp(ev, nil)

	if first != 1 {
		t.Fatalf("first handler expected 1 call, got %d", first)
	}
	if second != 2 {
		t.Fatalf("second handler expected 2 calls, got %d", second)
	}
}

// Listen's stop function removes the listener, releases the js.Func and closes
// the channel so consumer goroutines terminate.
func TestListenStop(t *testing.T) {
	doc := js.Document()
	el := doc.Call("createElement", "div")
	doc.Get("body").Call("appendChild", el)
	defer el.Call("remove")

	ch, stop := Listen("click", el)
	received := 0
	done := make(chan struct{})
	go func() {
		for range ch {
			received++
		}
		close(done)
	}()

	el.Call("click")
	// The dispatch is synchronous but delivery goes through the channel; give
	// the consumer goroutine a beat before stopping.
	time.Sleep(50 * time.Millisecond)
	stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("consumer goroutine did not terminate after stop")
	}
	if received != 1 {
		t.Fatalf("expected 1 event before stop, got %d", received)
	}
	// Events after stop must not fire the released handler (would panic on a
	// released js.Func if the listener were still attached).
	el.Call("click")
}
