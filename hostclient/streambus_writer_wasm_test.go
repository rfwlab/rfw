//go:build js && wasm

package hostclient

import (
	"strings"
	"testing"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

func streamBusWriterTest(t *testing.T, body string) (*hostConn, js.Value) {
	t.Helper()
	fixture := js.Window().Get("Function").New(`
		const state = {writes: []};
		state.writer = {write(value) {
			state.writes.push(value.length < 100 ? Array.from(value) : [value.length]);
			` + body + `
		}};
		return state;
	`).Invoke()
	c := newHostConn()
	startStreamBusWriter(c, &jsWriter{writer: fixture.Get("writer")})
	t.Cleanup(func() {
		c.fail(errConnectionClosed)
		if resolve := fixture.Get("resolve"); resolve.Type() == js.TypeFunction {
			resolve.Invoke()
		}
	})
	return c, fixture
}

func waitForStreamBusWrites(t *testing.T, fixture js.Value, count int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if fixture.Get("writes").Length() == count {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("stream writes = %d, want %d", fixture.Get("writes").Length(), count)
}

func TestStreamBusWriterReturnsInsideBrowserCallbackAndPreservesFrames(t *testing.T) {
	c, fixture := streamBusWriterTest(t, "return Promise.resolve();")
	var sendErr error
	callback := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		for _, data := range []string{"a", "bc", "def"} {
			if err := c.send(data); err != nil {
				sendErr = err
			}
		}
		return nil
	})
	defer callback.Release()
	callback.Invoke()
	if sendErr != nil {
		t.Fatalf("callback send: %v", sendErr)
	}
	waitForStreamBusWrites(t, fixture, 6)
	got := js.Window().Get("JSON").Call("stringify", fixture.Get("writes")).String()
	if got != "[[1],[97],[2],[98,99],[3],[100,101,102]]" {
		t.Fatalf("framed bytes = %s", got)
	}
}

func TestStreamBusWriterBoundsQueuedFrameCount(t *testing.T) {
	c, _ := streamBusWriterTest(t, "return new Promise(resolve => {state.resolve = resolve});")
	for i := 0; i < streamBusOutboundFrames; i++ {
		if err := c.send("x"); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}
	if err := c.send("x"); err == nil {
		t.Fatal("frame overflow accepted")
	}
	select {
	case <-c.done:
	default:
		t.Fatal("overflow left the connection open")
	}
}

func TestStreamBusWriterByteBudgetIncludesPendingNativeWrite(t *testing.T) {
	c, fixture := streamBusWriterTest(t, "return new Promise(resolve => {state.resolve = resolve});")
	if err := c.send(strings.Repeat("x", streamBusOutboundBytes)); err != nil {
		t.Fatalf("full budget: %v", err)
	}
	waitForStreamBusWrites(t, fixture, 1)
	if err := c.send("x"); err == nil {
		t.Fatal("in-flight bytes were omitted from the budget")
	}
}

func TestStreamBusWriterRefusesOversizedFrame(t *testing.T) {
	c, fixture := streamBusWriterTest(t, "return Promise.resolve();")
	if err := c.send(strings.Repeat("x", streamBusOutboundBytes+1)); err == nil {
		t.Fatal("oversized frame accepted")
	}
	if fixture.Get("writes").Length() != 0 {
		t.Fatal("oversized frame reached the native writer")
	}
}

func TestStreamBusWriterReleasesCompletedByteBudget(t *testing.T) {
	c, fixture := streamBusWriterTest(t, "return Promise.resolve();")
	data := strings.Repeat("x", streamBusOutboundBytes)
	if err := c.send(data); err != nil {
		t.Fatalf("first frame: %v", err)
	}
	waitForStreamBusWrites(t, fixture, 2)
	// The payload callback runs before its promise settles.
	time.Sleep(10 * time.Millisecond)
	if err := c.send(data); err != nil {
		t.Fatalf("completed frame retained its byte budget: %v", err)
	}
	waitForStreamBusWrites(t, fixture, 4)
}

func TestStreamBusWriterRejectedPromiseClosesConnection(t *testing.T) {
	c, _ := streamBusWriterTest(t, "return Promise.reject(new Error('write refused'));")
	if err := c.send("x"); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	select {
	case <-c.done:
	case <-time.After(time.Second):
		t.Fatal("rejected write left the connection open")
	}
	if err := c.send("y"); err == nil {
		t.Fatal("failed connection accepted another frame")
	}
}

func TestStreamBusWriterSynchronousExceptionClosesConnection(t *testing.T) {
	c, _ := streamBusWriterTest(t, "throw new Error('write refused');")
	if err := c.send("x"); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	select {
	case <-c.done:
	case <-time.After(time.Second):
		t.Fatal("throwing writer left the connection open")
	}
}

func TestStreamBusWriterCloseDoesNotFlushQueuedFrames(t *testing.T) {
	c, fixture := streamBusWriterTest(t, "return new Promise(resolve => {state.resolve = resolve});")
	if err := c.send("first"); err != nil {
		t.Fatalf("first frame: %v", err)
	}
	waitForStreamBusWrites(t, fixture, 1)
	if err := c.send("second"); err != nil {
		t.Fatalf("second frame: %v", err)
	}
	c.fail(errConnectionClosed)
	fixture.Get("resolve").Invoke()
	time.Sleep(20 * time.Millisecond)
	if n := fixture.Get("writes").Length(); n > 2 {
		t.Fatalf("writes after close = %d", n)
	}
	if err := c.send("third"); err == nil {
		t.Fatal("closed connection accepted another frame")
	}
}
