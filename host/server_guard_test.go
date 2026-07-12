package host

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func wsProbe(t *testing.T, mux *http.ServeMux, origin string) int {
	t.Helper()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/ws", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

// Without options the endpoint stays open: a plain GET reaches the WebSocket
// handler (which rejects the missing upgrade with 400, not a guard status).
func TestWSOpenByDefault(t *testing.T) {
	mux := NewMux(t.TempDir())
	if code := wsProbe(t, mux, "http://evil.example"); code == http.StatusForbidden || code == http.StatusUnauthorized {
		t.Fatalf("default mux rejected connection: %d", code)
	}
}

func TestWSOriginAllowlist(t *testing.T) {
	mux := NewMux(t.TempDir(), WithOriginAllowlist("https://app.example.com"))
	if code := wsProbe(t, mux, "http://evil.example"); code != http.StatusForbidden {
		t.Fatalf("expected 403 for unlisted origin, got %d", code)
	}
	if code := wsProbe(t, mux, ""); code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing origin, got %d", code)
	}
	if code := wsProbe(t, mux, "https://app.example.com"); code == http.StatusForbidden {
		t.Fatalf("allowed origin rejected: %d", code)
	}
}

func TestWSAuthFunc(t *testing.T) {
	mux := NewMux(t.TempDir(), WithAuthFunc(func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://trusted.example.com"
	}))
	if code := wsProbe(t, mux, "http://evil.example"); code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for rejected auth, got %d", code)
	}
	if code := wsProbe(t, mux, "https://trusted.example.com"); code == http.StatusUnauthorized {
		t.Fatalf("accepted auth rejected: %d", code)
	}
}
