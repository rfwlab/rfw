package host

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// buildDir writes a client build with the artifacts the test names and returns
// the mux serving it.
func buildDir(t *testing.T, artifacts map[string]string) (*http.ServeMux, map[string]int) {
	t.Helper()
	dir := t.TempDir()
	client := filepath.Join(dir, "client")
	if err := os.MkdirAll(client, 0o750); err != nil {
		t.Fatalf("mkdir client: %v", err)
	}
	sizes := map[string]int{}
	for name, body := range artifacts {
		if err := os.WriteFile(filepath.Join(client, name), []byte(body), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		sizes[name] = len(body)
	}
	if err := os.WriteFile(filepath.Join(client, "index.html"), []byte("<html></html>"), 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}
	return NewMux(client), sizes
}

func request(t *testing.T, mux *http.ServeMux, target, acceptEncoding string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if acceptEncoding != "" {
		req.Header.Set("Accept-Encoding", acceptEncoding)
	}
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)
	return recorder.Result()
}

var allArtifacts = map[string]string{
	"app.wasm":    "raw wasm bytes, the big one",
	"app.wasm.br": "brotli",
	"app.wasm.gz": "gzipped",
}

// A client that accepts brotli gets brotli from the raw URL, with the headers
// that let the browser decode it and a cache entry a shared cache cannot
// mis-serve.
func TestNegotiationPrefersBrotli(t *testing.T) {
	mux, sizes := buildDir(t, allArtifacts)
	resp := request(t, mux, "/app.wasm?v=abc123", "gzip, deflate, br")
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Content-Encoding"); got != "br" {
		t.Fatalf("Content-Encoding = %q, want br", got)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/wasm" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := resp.Header.Get("Content-Length"); got != strconv.Itoa(sizes["app.wasm.br"]) {
		t.Fatalf("Content-Length = %q, want the compressed size %d", got, sizes["app.wasm.br"])
	}
	if got := resp.Header.Get("Vary"); got != "Accept-Encoding" {
		t.Fatalf("Vary = %q", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control on a versioned URL = %q", got)
	}
}

// A plain HTTP browser advertises gzip but not brotli, and must get gzip
// rather than the raw bundle.
func TestNegotiationFallsBackToGzip(t *testing.T) {
	mux, sizes := buildDir(t, allArtifacts)
	resp := request(t, mux, "/app.wasm", "gzip, deflate")
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	if got := resp.Header.Get("Content-Length"); got != strconv.Itoa(sizes["app.wasm.gz"]) {
		t.Fatalf("Content-Length = %q, want %d", got, sizes["app.wasm.gz"])
	}
	// An unversioned URL must revalidate or a release cannot replace it.
	if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("Cache-Control on an unversioned URL = %q", got)
	}
}

// Brotli is never sent to a client that did not advertise it, whatever the
// build produced.
func TestNegotiationNeverSendsUnrequestedBrotli(t *testing.T) {
	mux, _ := buildDir(t, allArtifacts)
	for _, accept := range []string{"gzip", "gzip, deflate", "identity", "br;q=0, gzip"} {
		resp := request(t, mux, "/app.wasm", accept)
		encoding := resp.Header.Get("Content-Encoding")
		_ = resp.Body.Close()
		if encoding == "br" {
			t.Fatalf("Accept-Encoding %q got brotli", accept)
		}
	}
}

// q=0 is an explicit refusal, not a low preference.
func TestNegotiationHonoursAnExplicitRefusal(t *testing.T) {
	mux, _ := buildDir(t, allArtifacts)
	resp := request(t, mux, "/app.wasm", "br;q=0, gzip;q=0")
	defer func() { _ = resp.Body.Close() }()
	if got := resp.Header.Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding = %q, want none", got)
	}
}

// A client that sends no Accept-Encoding gets the raw bundle, which is the
// only body it is guaranteed to understand.
func TestNegotiationServesRawWithoutAnAcceptEncoding(t *testing.T) {
	mux, sizes := buildDir(t, allArtifacts)
	resp := request(t, mux, "/app.wasm", "")
	defer func() { _ = resp.Body.Close() }()
	if got := resp.Header.Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding = %q, want none", got)
	}
	if got := resp.Header.Get("Content-Length"); got != strconv.Itoa(sizes["app.wasm"]) {
		t.Fatalf("Content-Length = %q, want the raw size", got)
	}
}

// A build with no compressed artifact still serves, so a development build
// keeps working.
func TestNegotiationFallsThroughWhenNoArtifactExists(t *testing.T) {
	mux, _ := buildDir(t, map[string]string{"app.wasm": "raw only"})
	resp := request(t, mux, "/app.wasm", "gzip, br")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding = %q with no artifact on disk", got)
	}
}

// Only gzip exists, and a brotli-capable client takes it.
func TestNegotiationSkipsAMissingBrotliArtifact(t *testing.T) {
	mux, _ := buildDir(t, map[string]string{"app.wasm": "raw", "app.wasm.gz": "gzipped"})
	resp := request(t, mux, "/app.wasm", "br, gzip")
	defer func() { _ = resp.Body.Close() }()
	if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
}

// A directly addressed artifact, which is how static hosting reaches it, is
// still labelled correctly.
func TestDirectArtifactURLsAreLabelled(t *testing.T) {
	mux, _ := buildDir(t, allArtifacts)
	for path, want := range map[string]string{
		"/app.wasm.br": "br",
		"/app.wasm.gz": "gzip",
	} {
		resp := request(t, mux, path+"?v=abc123", "gzip, br")
		encoding := resp.Header.Get("Content-Encoding")
		contentType := resp.Header.Get("Content-Type")
		cache := resp.Header.Get("Cache-Control")
		vary := resp.Header.Get("Vary")
		_ = resp.Body.Close()
		if encoding != want {
			t.Errorf("%s Content-Encoding = %q, want %q", path, encoding, want)
		}
		if contentType != "application/wasm" {
			t.Errorf("%s Content-Type = %q", path, contentType)
		}
		if cache != "public, max-age=31536000, immutable" {
			t.Errorf("%s Cache-Control = %q", path, cache)
		}
		if vary != "Accept-Encoding" {
			t.Errorf("%s Vary = %q", path, vary)
		}
	}
}

// The pointer file has to revalidate, or a browser keeps an old version and
// therefore an old bundle URL.
func TestClientConfigRevalidates(t *testing.T) {
	mux, _ := buildDir(t, map[string]string{"app.wasm": "raw", "rfw_config.js": "window.RFW_WASM_VERSION = \"abc\";"})
	resp := request(t, mux, "/rfw_config.js?v=abc123", "gzip")
	defer func() { _ = resp.Body.Close() }()
	if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("rfw_config.js Cache-Control = %q, want no-cache", got)
	}
}

// index.html carries the stamped bootstrap script tags, so a browser that
// caches it keeps loading the previous release's loader. It has to revalidate.
func TestStampedPageRevalidates(t *testing.T) {
	mux, _ := buildDir(t, map[string]string{"app.wasm": "raw"})
	for _, target := range []string{"/", "/index.html"} {
		resp := request(t, mux, target, "gzip")
		cache := resp.Header.Get("Cache-Control")
		_ = resp.Body.Close()
		if cache != "no-cache" {
			t.Errorf("%s Cache-Control = %q, want no-cache", target, cache)
		}
	}
}

func TestAcceptedEncodings(t *testing.T) {
	cases := map[string]map[string]float64{
		"":                     nil,
		"gzip":                 {"gzip": 1},
		"gzip, deflate, br":    {"gzip": 1, "deflate": 1, "br": 1},
		"BR":                   {"br": 1},
		" gzip ;q=0.5 , br ":   {"gzip": 0.5, "br": 1},
		"br;q=0, gzip":         {"br": 0, "gzip": 1},
		"*":                    {"*": 1},
		"*, br;q=0":            {"*": 1, "br": 0},
		"identity;q=1, br;q=0": {"identity": 1, "br": 0},
		"br;q=bogus, gzip":     {"br": 0, "gzip": 1},
		"br;q=2, gzip":         {"br": 0, "gzip": 1},
	}
	for header, want := range cases {
		got := acceptedEncodings(header)
		if len(got) != len(want) {
			t.Errorf("acceptedEncodings(%q) = %v, want %v", header, got, want)
			continue
		}
		for name, quality := range want {
			if got[name] != quality {
				t.Errorf("acceptedEncodings(%q)[%q] = %v, want %v", header, name, got[name], quality)
			}
		}
	}
}

func TestNegotiationHonorsClientEncodingPreference(t *testing.T) {
	mux, _ := buildDir(t, map[string]string{
		"app.wasm":    "raw",
		"app.wasm.br": "brotli",
		"app.wasm.gz": "gzip",
	})
	resp := request(t, mux, "/app.wasm", "gzip;q=1, br;q=0.1")
	defer func() { _ = resp.Body.Close() }()
	if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
}

func TestNegotiationUsesServerPreferenceForEqualQualities(t *testing.T) {
	mux, _ := buildDir(t, map[string]string{
		"app.wasm":    "raw",
		"app.wasm.br": "brotli",
		"app.wasm.gz": "gzip",
	})
	resp := request(t, mux, "/app.wasm", "gzip, br")
	defer func() { _ = resp.Body.Close() }()
	if got := resp.Header.Get("Content-Encoding"); got != "br" {
		t.Fatalf("Content-Encoding = %q, want br", got)
	}
}
