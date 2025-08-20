//go:build js && wasm

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rfwlab/rfw/docs/examples/components"
)

func TestFetchData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	data, err := components.FetchData(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data != "ok" {
		t.Fatalf("expected 'ok', got %s", data)
	}
}
