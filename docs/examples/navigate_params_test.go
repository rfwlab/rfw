//go:build js && wasm

package main

import (
	"testing"

	"github.com/rfwlab/rfw/docs/examples/components"
	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/router"
)

func TestNavigateMergesParamsAndQuery(t *testing.T) {
	router.Reset()
	c := components.NewParamsComponent()
	router.RegisterRoute(router.Route{
		Path:      "/examples/params/:id",
		Component: func() core.Component { return c },
	})
	router.Navigate("/examples/params/42?tab=posts")
	if c.ID != "42" || c.Tab != "posts" {
		t.Fatalf("expected params merged, got id=%s tab=%s", c.ID, c.Tab)
	}
}
