package ssc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rfwlab/rfw/v2/host"
)

func TestSSCEventBus(t *testing.T) {
	type seenEvent struct {
		component string
		value     any
	}
	seen := make(chan seenEvent, 1)
	SubscribeSSC(func(ctx context.Context, e SSCEvent) error {
		seen <- seenEvent{component: e.Component, value: e.Payload["value"]}
		return nil
	})

	if err := EmitSSC(context.Background(), SSCEvent{Component: "Counter", Payload: map[string]any{"value": 2}}); err != nil {
		t.Fatalf("emit failed: %v", err)
	}

	got := <-seen
	if got.component != "Counter" || got.value != 2 {
		t.Fatalf("unexpected event: %+v", got)
	}
}

func TestSSCServerServesIndexAndWasmHeaders(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<main>app</main>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.wasm.br"), []byte("wasm"), 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}

	server := NewSSCServer(":0", root)
	ts := httptest.NewServer(server.Mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/docs/anything")
	if err != nil {
		t.Fatalf("index fallback request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected index fallback 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/app.wasm.br")
	if err != nil {
		t.Fatalf("wasm request failed: %v", err)
	}
	resp.Body.Close()
	if resp.Header.Get("Content-Encoding") != "br" {
		t.Fatalf("expected br encoding, got %q", resp.Header.Get("Content-Encoding"))
	}
	if resp.Header.Get("Content-Type") != "application/wasm" {
		t.Fatalf("expected wasm content type, got %q", resp.Header.Get("Content-Type"))
	}
}

func TestSSCWithSessionTargetDelegatesHostOption(t *testing.T) {
	var opts host.BroadcastOptions
	WithSessionTarget("abc")(&opts)
	if opts.Session != "abc" {
		t.Fatalf("expected session abc, got %q", opts.Session)
	}
}
