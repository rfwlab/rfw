//go:build js && wasm

package main

import (
	"strings"
	"testing"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/state"
)

func TestEventBindingToken(t *testing.T) {
	tpl := []byte("<root><button @click:increment></button></root>")
	store := state.NewStore("default")
	c := core.NewHTMLComponent("TestComponent", tpl, nil)
	c.Init(store)
	html := c.Render()
	if !strings.Contains(html, "data-on-click=\"increment\"") {
		t.Fatalf("expected data-on-click attribute, got %s", html)
	}
}
