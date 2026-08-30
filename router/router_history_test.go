//go:build js && wasm

package router

import (
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/core"
	events "github.com/rfwlab/rfw/v2/events"
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

// restoreHistoryPath puts the page back on the path the test found, so the
// suite that follows starts where it expects to.
func restoreHistoryPath(t *testing.T) {
	t.Helper()
	original := js.Location().Get("pathname").String() +
		js.Location().Get("search").String() +
		js.Location().Get("hash").String()
	t.Cleanup(func() {
		js.History().Call("replaceState", nil, "", original)
		Reset()
	})
}

// travelHistory triggers a browser history move and waits for the popstate
// that carries the new URL.
func travelHistory(t *testing.T, direction string) string {
	t.Helper()
	popstate, stop := events.Listen("popstate", js.Window())
	defer stop()
	js.History().Call(direction)
	select {
	case <-popstate:
	case <-time.After(2 * time.Second):
		t.Fatalf("history %s did not fire popstate", direction)
	}
	return js.Location().Get("pathname").String()
}

// A login redirect replaces the entry it was serving, so the protected path
// leaves no history entry to go back to.
func TestGuardReplaceLeavesNoHistoryEntry(t *testing.T) {
	Reset()
	restoreHistoryPath(t)

	RegisterRoute(Route{Path: "/guard-base", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{Path: "/guard-entry", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{Path: "/guard-login", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{
		Path:         "/guard-account",
		Component:    func() core.Component { return routeComponent{} },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return ReplaceWith("/guard-login") }},
	})

	Navigate("/guard-base")
	Navigate("/guard-entry")
	before := js.History().Get("length").Int()

	Navigate("/guard-account")

	if got := js.Location().Get("pathname").String(); got != "/guard-login" {
		t.Fatalf("path after the login redirect = %q", got)
	}
	if got := js.History().Get("length").Int(); got != before {
		t.Fatalf("history length changed from %d to %d", before, got)
	}
	if got := travelHistory(t, "back"); got != "/guard-base" {
		t.Fatalf("back landed on %q, want /guard-base", got)
	}
	if got := travelHistory(t, "forward"); got != "/guard-login" {
		t.Fatalf("forward landed on %q, want /guard-login", got)
	}
}

// A push redirect keeps the entry it came from, so back returns to it.
func TestGuardRedirectPushesHistoryEntry(t *testing.T) {
	Reset()
	restoreHistoryPath(t)

	RegisterRoute(Route{Path: "/guard-origin", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{Path: "/guard-offers", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{
		Path:         "/guard-promo",
		Component:    func() core.Component { return routeComponent{} },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return RedirectTo("/guard-offers") }},
	})

	Navigate("/guard-origin")
	before := js.History().Get("length").Int()

	Navigate("/guard-promo")

	if got := js.Location().Get("pathname").String(); got != "/guard-offers" {
		t.Fatalf("path after the guard redirect = %q", got)
	}
	if got := js.History().Get("length").Int(); got != before+1 {
		t.Fatalf("history length = %d, want %d", got, before+1)
	}
	if got := travelHistory(t, "back"); got != "/guard-origin" {
		t.Fatalf("back landed on %q, want /guard-origin", got)
	}
	if got := travelHistory(t, "forward"); got != "/guard-offers" {
		t.Fatalf("forward landed on %q, want /guard-offers", got)
	}
}

// Landing directly on a protected path, the way a deep link does, ends on the
// login route without committing the protected one.
func TestGuardReplaceOnDeepLink(t *testing.T) {
	Reset()
	restoreHistoryPath(t)

	created := 0
	RegisterRoute(Route{Path: "/deep-login", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{
		Path: "/deep-account",
		Component: func() core.Component {
			created++
			return routeComponent{}
		},
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return ReplaceWith("/deep-login") }},
	})

	js.History().Call("pushState", nil, "", "/deep-account")
	navigate("/deep-account", historyNone)

	if got := js.Location().Get("pathname").String(); got != "/deep-login" {
		t.Fatalf("deep link path = %q, want /deep-login", got)
	}
	if created != 0 {
		t.Fatalf("protected component was created %d times", created)
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
