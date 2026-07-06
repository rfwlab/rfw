package host

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMuxDebugEndpoints(t *testing.T) {
	t.Setenv("RFW_DEVTOOLS", "1")
	mux := NewMux(t.TempDir())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/debug/vars")
	if err != nil {
		t.Fatalf("vars request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/debug/pprof/")
	if err != nil {
		t.Fatalf("pprof request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
