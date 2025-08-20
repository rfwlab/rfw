//go:build js && wasm

package main

import (
	"strings"
	"testing"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/state"
)

// TestStorePlaceholderPreserved ensures that store bindings inside
// attribute values are not mistaken for event directives.
func TestStorePlaceholderPreserved(t *testing.T) {
	tpl := []byte("<root><input value=\"@store:default.testing.testingState:w\"></root>")
	// Component store
	store := state.NewStore("default")
	// Store referenced by the placeholder
	state.NewStore("testing")
	c := core.NewHTMLComponent("TestComponent", tpl, nil)
	c.Init(store)
	html := c.Render()
	if strings.Contains(html, "data-on-store") {
		t.Fatalf("unexpected data-on-store attribute: %s", html)
	}
	if !strings.Contains(html, "value=\"@store:default.testing.testingState:w\"") {
		t.Fatalf("store placeholder was altered: %s", html)
	}
}
