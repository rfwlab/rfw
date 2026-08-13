//go:build !js

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
		"rfw_config.js":  {Data: []byte("//cfg")},
		"assets/app.css": {Data: []byte(".a{}")},
	}
}

func TestNewMuxFSServesEmbeddedBuild(t *testing.T) {
	srv := httptest.NewServer(NewMuxFS(testClientFS()))
	defer srv.Close()

	type result struct {
		status int
		header http.Header
	}
	// get issues the request and closes the body before returning, so the
	// assertions never hold an open response.
	get := func(path, accept string) result {
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
		if cerr := resp.Body.Close(); cerr != nil {
			t.Fatalf("close body %s: %v", path, cerr)
		}
		return result{status: resp.StatusCode, header: resp.Header}
	}

	if got := get("/", "text/html").status; got != http.StatusOK {
		t.Fatalf("root status = %d, want 200", got)
	}
	if got := get("/assets/app.css", "").status; got != http.StatusOK {
		t.Fatalf("nested asset status = %d, want 200", got)
	}
	if got := get("/app.wasm?v=abc", "").header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("versioned wasm Cache-Control = %q", got)
	}
	if got := get("/app.wasm.br", "").header.Get("Content-Encoding"); got != "br" {
		t.Fatalf("wasm.br Content-Encoding = %q, want br", got)
	}
	if got := get("/rfw_config.js", "").header.Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("rfw_config.js Cache-Control = %q, want no-cache", got)
	}
	// An unknown HTML route falls back to index.html (single-page app).
	if got := get("/dashboard/live", "text/html").status; got != http.StatusOK {
		t.Fatalf("unknown html route status = %d, want 200 (index fallback)", got)
	}
	// A missing non-HTML asset is a 404, not the index.
	if got := get("/missing.css", "").status; got != http.StatusNotFound {
		t.Fatalf("missing asset status = %d, want 404", got)
	}
	// A plain GET is not a WebSocket handshake, so /ws rejects it, but it must be
	// routed rather than falling through to the catch-all 404.
	if got := get("/ws", "").status; got == http.StatusNotFound {
		t.Fatalf("/ws returned 404, endpoint not registered")
	}
}
