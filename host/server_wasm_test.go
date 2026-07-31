//go:build !js

package host

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewMuxServesBrotliWasmWithEncoding(t *testing.T) {
	t.Setenv("RFW_DEVTOOLS", "")
	root := t.TempDir()
	clientDir := filepath.Join(root, "client")
	if err := os.MkdirAll(clientDir, 0o755); err != nil {
		t.Fatalf("failed to create client dir: %v", err)
	}
	wasmPath := filepath.Join(clientDir, "app.wasm.br")
	if err := os.WriteFile(wasmPath, []byte("compressed"), 0o644); err != nil {
		t.Fatalf("failed to write wasm: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clientDir, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatalf("failed to write index: %v", err)
	}

	mux := NewMux(clientDir)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/app.wasm.br?v=abc123")
	if err != nil {
		t.Fatalf("failed to get wasm: %v", err)
	}
	defer closeTestResource(t, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if enc := resp.Header.Get("Content-Encoding"); enc != "br" {
		t.Fatalf("expected Content-Encoding br, got %q", enc)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/wasm" {
		t.Fatalf("expected Content-Type application/wasm, got %q", ct)
	}
	if cache := resp.Header.Get("Cache-Control"); cache != "public, max-age=31536000, immutable" {
		t.Fatalf("unexpected Cache-Control header: %q", cache)
	}
	if vary := resp.Header.Get("Vary"); vary != "Accept-Encoding" && vary != "Accept-Encoding, Accept-Encoding" {
		// Allow duplicated value as Go's header may append values depending on environment.
		t.Fatalf("unexpected Vary header: %q", vary)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body) != "compressed" {
		t.Fatalf("unexpected body: %q", string(body))
	}

	resp, err = http.Get(srv.URL + "/app.wasm.br?v=")
	if err != nil {
		t.Fatalf("failed to get unversioned wasm: %v", err)
	}
	defer closeTestResource(t, resp.Body)
	if cache := resp.Header.Get("Cache-Control"); cache != "no-cache" {
		t.Fatalf("unexpected unversioned Cache-Control header: %q", cache)
	}
}
