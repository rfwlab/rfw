//go:build js && wasm

package router

import (
	"testing"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/js"
)

func TestReplaceKeepsHistoryLength(t *testing.T) {
	Reset()
	originalPath := js.Location().Get("pathname").String() +
		js.Location().Get("search").String() +
		js.Location().Get("hash").String()
	t.Cleanup(func() {
		js.History().Call("replaceState", nil, "", originalPath)
		Reset()
	})

	RegisterRoute(Route{
		Path:      "/replace-history",
		Component: func() core.Component { return routeComponent{} },
	})
	before := js.History().Get("length").Int()

	Replace("/replace-history")

	if got := js.Location().Get("pathname").String(); got != "/replace-history" {
		t.Fatalf("expected replaced path, got %q", got)
	}
	if got := js.History().Get("length").Int(); got != before {
		t.Fatalf("history length changed from %d to %d", before, got)
	}
}

func TestHistoryNoneGuardFallbackDoesNotPush(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	RegisterRoute(Route{
		Path:      "/",
		Component: func() core.Component { return routeComponent{} },
	})
	RegisterRoute(Route{
		Path:      "/protected",
		Component: func() core.Component { return routeComponent{} },
		Guards:    []Guard{func(map[string]string) bool { return false }},
	})
	before := js.History().Get("length").Int()

	navigate("/protected", historyNone)

	if got := js.History().Get("length").Int(); got != before {
		t.Fatalf("history length changed from %d to %d", before, got)
	}
}
