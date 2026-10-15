package host

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadPortOverride(t *testing.T) {
	t.Setenv("RFW_HOST_PORT", "9095")
	if got := readPort(); got != 9095 {
		t.Fatalf("expected override port 9095, got %d", got)
	}
}

func TestStreamBusKeepsConfiguredOriginAndIdentityGuards(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	mux := NewMux(t.TempDir(), WithOriginAllowlist("https://allowed.example"), WithAuthFunc(func(r *http.Request) bool { return r.Header.Get("Authorization") == "valid" }))
	request := httptest.NewRequest("GET", "https://host/streambus", nil)
	request.Header.Set("Origin", "https://denied.example")
	request.Header.Set("Authorization", "valid")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("denied origin returned %d", response.Code)
	}
	request.Header.Set("Origin", "https://allowed.example")
	request.Header.Del("Authorization")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing identity returned %d", response.Code)
	}
	request.Header.Set("Authorization", "valid")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusUpgradeRequired {
		t.Fatalf("allowed request returned %d", response.Code)
	}
}
