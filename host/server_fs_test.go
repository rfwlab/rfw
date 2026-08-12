package host

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func testClientFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":     {Data: []byte("<!doctype html><div id=app></div>")},
		"app.wasm":       {Data: []byte("\x00asm")},
		"app.wasm.br":    {Data: []byte("brotli-bytes")},
		"assets/app.css": {Data: []byte(".a{}")},
	}
}

func TestNewMuxFSServesEmbeddedBuild(t *testing.T) {
	srv := httptest.NewServer(NewMuxFS(testClientFS()))
	defer srv.Close()

	get := func(path string, accept string) *http.Response {
		req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
		if err != nil {
			t.Fatalf("new request %s: %v", path, err)
		}
		if accept != "" {
			req.Header.Set("Accept", accept)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("get %s: %v", path, err)
		}
		return resp
	}

	t.Run("index at root", func(t *testing.T) {
		resp := get("/", "text/html")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("root status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("nested asset", func(t *testing.T) {
		resp := get("/assets/app.css", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("asset status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("wasm cache header", func(t *testing.T) {
		resp := get("/app.wasm?v=abc", "")
		defer resp.Body.Close()
		if got := resp.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("versioned wasm Cache-Control = %q", got)
		}
	})

	t.Run("brotli wasm encoding", func(t *testing.T) {
		resp := get("/app.wasm.br", "")
		defer resp.Body.Close()
		if got := resp.Header.Get("Content-Encoding"); got != "br" {
			t.Fatalf("wasm.br Content-Encoding = %q, want br", got)
		}
	})

	t.Run("SPA fallback for unknown html route", func(t *testing.T) {
		resp := get("/dashboard/live", "text/html")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unknown html route status = %d, want 200 (index fallback)", resp.StatusCode)
		}
	})

	t.Run("missing non-html asset is 404", func(t *testing.T) {
		resp := get("/missing.css", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("missing asset status = %d, want 404", resp.StatusCode)
		}
	})

	t.Run("ws endpoint is registered", func(t *testing.T) {
		resp := get("/ws", "")
		defer resp.Body.Close()
		// A plain GET is not a WebSocket handshake, so the handler rejects it,
		// but it must be routed (not the catch-all 404).
		if resp.StatusCode == http.StatusNotFound {
			t.Fatalf("/ws returned 404, endpoint not registered")
		}
	})
}
