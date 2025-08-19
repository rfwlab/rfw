//go:build js && wasm

package main

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/example/components"
)

func TestComplexRoutingProps(t *testing.T) {
	c := components.NewComplexRoutingComponent()
	c.SetRouteParams(map[string]string{"user": "alice", "section": "settings"})
	html := c.Render()
	if !strings.Contains(html, "alice") || !strings.Contains(html, "settings") {
		t.Fatalf("expected route params in HTML, got %s", html)
	}
}
