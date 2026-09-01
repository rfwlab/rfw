//go:build js && wasm

package hostclient

import (
	"context"
	"strings"
	"testing"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

const fakeCapacitorSSCSource = `
(function () {
  window.__fakeRFWSSC = {
    connects: [],
    sends: [],
    closes: [],
    closeDelayMs: -1,
    terminalCallbacks: 0,
    callback: null,
    connect: function (options, callback) {
      this.connects.push(options);
      this.callback = callback;
      callback({ type: "open" });
    },
    send: function (options) { this.sends.push(options); },
    close: function (options) {
      var self = this;
      this.closes.push(options);
      if (this.closeDelayMs >= 0) {
        setTimeout(function () {
          self.terminalCallbacks++;
          self.callback({ type: "close", code: 1000, reason: "closed" });
        }, this.closeDelayMs);
      }
    }
  };
  window.Capacitor = { Plugins: { RFWSSC: window.__fakeRFWSSC } };
})();
`

func installFakeCapacitorSSC(t *testing.T) js.Value {
	t.Helper()
	originalCapacitor := js.Get("Capacitor")
	originalTransport := js.Get("RFW_SSC_TRANSPORT")
	js.Call("eval", fakeCapacitorSSCSource)
	js.Set("RFW_SSC_TRANSPORT", sscTransportCapacitor)
	t.Cleanup(func() {
		js.Set("Capacitor", originalCapacitor)
		js.Set("RFW_SSC_TRANSPORT", originalTransport)
		js.Global().Delete("__fakeRFWSSC")
	})
	return js.Get("__fakeRFWSSC")
}

func TestSSCTransportDefaultsToBrowser(t *testing.T) {
	original := js.Get("RFW_SSC_TRANSPORT")
	js.Global().Delete("RFW_SSC_TRANSPORT")
	t.Cleanup(func() { js.Set("RFW_SSC_TRANSPORT", original) })

	if got := sscTransport(); got != sscTransportBrowser {
		t.Fatalf("transport = %q, want browser", got)
	}
}

func TestCapacitorTransportFailsClosedWhenPluginIsMissing(t *testing.T) {
	originalCapacitor := js.Get("Capacitor")
	originalTransport := js.Get("RFW_SSC_TRANSPORT")
	js.Set("Capacitor", js.NewDict().Value)
	js.Set("RFW_SSC_TRANSPORT", sscTransportCapacitor)
	t.Cleanup(func() {
		js.Set("Capacitor", originalCapacitor)
		js.Set("RFW_SSC_TRANSPORT", originalTransport)
	})

	_, err := dial(context.Background(), "wss://api.example.com/ws")
	if err == nil || !strings.Contains(err.Error(), "plugins are unavailable") {
		t.Fatalf("dial error = %v, want missing-plugin failure", err)
	}
}

func TestCapacitorTransportConnectsWritesReadsAndCloses(t *testing.T) {
	plugin := installFakeCapacitorSSC(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := dial(ctx, "wss://api.example.com/ws")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if got := plugin.Get("connects").Get("length").Int(); got != 1 {
		t.Fatalf("connect calls = %d, want 1", got)
	}
	connect := plugin.Get("connects").Index(0)
	if got := connect.Get("url").String(); got != "wss://api.example.com/ws" {
		t.Fatalf("connect url = %q", got)
	}
	if connect.Get("id").String() == "" {
		t.Fatal("connect id is empty")
	}

	if err := conn.writeJSON(wireMessage{Component: "ticker", Sequence: 1}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := plugin.Get("sends").Get("length").Int(); got != 1 {
		t.Fatalf("send calls = %d, want 1", got)
	}
	if frame := plugin.Get("sends").Index(0).Get("data").String(); !strings.Contains(frame, `"component":"ticker"`) {
		t.Fatalf("sent frame = %q", frame)
	}

	event := js.NewDict()
	event.Set("type", "message")
	event.Set("encoding", "text")
	event.Set("data", `{"component":"ticker","sequence":1}`)
	plugin.Get("callback").Invoke(event.Value)
	frame, err := conn.read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got := string(frame); got != `{"component":"ticker","sequence":1}` {
		t.Fatalf("frame = %q", got)
	}

	if err := conn.close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if got := plugin.Get("closes").Get("length").Int(); got != 1 {
		t.Fatalf("close calls = %d, want 1", got)
	}
}

func TestCapacitorTransportRejectsMalformedBinaryFrame(t *testing.T) {
	plugin := installFakeCapacitorSSC(t)
	conn, err := dial(context.Background(), "wss://api.example.com/ws")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.close() })

	event := js.NewDict()
	event.Set("type", "message")
	event.Set("encoding", "base64")
	event.Set("data", "not base64")
	plugin.Get("callback").Invoke(event.Value)

	if _, err := conn.read(context.Background()); err == nil || !strings.Contains(err.Error(), "invalid base64") {
		t.Fatalf("read error = %v, want invalid base64", err)
	}
}

func TestCapacitorTransportKeepsCallbackAliveThroughAsyncClose(t *testing.T) {
	plugin := installFakeCapacitorSSC(t)
	plugin.Set("closeDelayMs", 10)
	conn, err := dial(context.Background(), "wss://api.example.com/ws")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	if err := conn.close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := conn.read(ctx); err == nil || !strings.Contains(err.Error(), "code 1000") {
		t.Fatalf("read error = %v, want native close", err)
	}
	if got := plugin.Get("terminalCallbacks").Int(); got != 1 {
		t.Fatalf("terminal callbacks = %d, want 1", got)
	}
}

func TestCapacitorTransportDoesNotFallBackForUnknownMode(t *testing.T) {
	original := js.Get("RFW_SSC_TRANSPORT")
	js.Set("RFW_SSC_TRANSPORT", "native")
	t.Cleanup(func() { js.Set("RFW_SSC_TRANSPORT", original) })

	_, err := dial(context.Background(), "wss://api.example.com/ws")
	if err == nil || !strings.Contains(err.Error(), "unsupported SSC transport") {
		t.Fatalf("dial error = %v, want unsupported transport", err)
	}
}
