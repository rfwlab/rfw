//go:build js && wasm

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rfwhttp "github.com/rfwlab/rfw/v1/http"
)

func TestFetchJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"title":"ok"}`))
	}))
	defer srv.Close()

	var todo struct {
		Title string `json:"title"`
	}
	if err := rfwhttp.FetchJSON(srv.URL, &todo); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if todo.Title != "ok" {
		t.Fatalf("expected 'ok', got %s", todo.Title)
	}
}
