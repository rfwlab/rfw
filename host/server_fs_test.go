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

func TestNewMuxFSNegotiatesTheSameCompressedBundleAsDirectoryServing(t *testing.T) {
	t.Setenv("RFW_DEV_BUILD", "")
	mux := NewMuxFS(testClientFS())
	request := httptest.NewRequest(http.MethodGet, "/app.wasm?v=build1", nil)
	request.Header.Set("Accept-Encoding", "br, gzip")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != 200 || response.Body.String() != "brotli-bytes" {
		t.Fatalf("negotiated response: code=%d body=%q", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Encoding") != "br" || response.Header().Get("Vary") != "Accept-Encoding" || response.Header().Get("Content-Type") != "application/wasm" {
		t.Fatalf("negotiated headers: %v", response.Header())
	}
	if response.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" || response.Header().Get("Content-Length") != "12" {
		t.Fatalf("versioned artifact headers: %v", response.Header())
	}
	request = httptest.NewRequest(http.MethodHead, "/app.wasm", nil)
	request.Header.Set("Accept-Encoding", "br")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != 200 || response.Body.Len() != 0 || response.Header().Get("Content-Length") != "12" || response.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("HEAD response: code=%d body=%q headers=%v", response.Code, response.Body.String(), response.Header())
	}
	request = httptest.NewRequest(http.MethodGet, "/app.wasm", nil)
	request.Header.Set("Accept-Encoding", "br;q=0, gzip;q=0")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Header().Get("Content-Encoding") != "" || response.Body.String() != "\x00asm" {
		t.Fatalf("identity response: headers=%v body=%q", response.Header(), response.Body.String())
	}
}

func TestNewMuxFSRevalidatesSPAEntriesAndDisablesDevelopmentCaching(t *testing.T) {
	t.Setenv("RFW_DEV_BUILD", "")
	mux := NewMuxFS(testClientFS())
	request := httptest.NewRequest(http.MethodGet, "/nested/route", nil)
	request.Header.Set("Accept", "text/html")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != 200 || response.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("SPA entry: code=%d headers=%v", response.Code, response.Header())
	}
	t.Setenv("RFW_DEV_BUILD", "1")
	request = httptest.NewRequest(http.MethodGet, "/", nil)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != 200 || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("development entry: code=%d headers=%v", response.Code, response.Header())
	}
	request = httptest.NewRequest(http.MethodGet, "/app.wasm?v=old", nil)
	request.Header.Set("Accept-Encoding", "br")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("development artifact: headers=%v", response.Header())
	}
}

func TestNewMuxFSStreamBusUsesTheSameOriginAndIdentityGuard(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	mux := NewMuxFS(testClientFS(), WithOriginAllowlist("https://allowed.example"), WithAuthFunc(func(r *http.Request) bool { return r.Header.Get("Authorization") == "valid" }))
	t.Cleanup(func() { streamEndpoints.Delete(mux) })
	request := httptest.NewRequest(http.MethodConnect, streamBusPath, nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("missing origin: %d", response.Code)
	}
	request.Header.Set("Origin", "https://allowed.example")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing identity: %d", response.Code)
	}
	request.Header.Set("Authorization", "valid")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusUpgradeRequired {
		t.Fatalf("guarded non-HTTP3 request: %d", response.Code)
	}
}

func TestNewMuxFSStreamBusWebTransportProtocol(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	testStreamBusMux(t, NewMuxFS(testClientFS()))
}
