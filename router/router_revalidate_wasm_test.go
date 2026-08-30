//go:build js && wasm

package router

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/state"
)

// mountRevalidateShell renders a shell with a router outlet into #app and
// mounts page behind guard on path. page must carry the "reval" marker, which
// is filled after mount the way a fetch callback fills a card: a remount or a
// stray re-render would wipe it. It returns the shell so a caller can hold on
// to the root MountRoot mounted.
func mountRevalidateShell(t *testing.T, path string, page *hostGuardPage, guard ResultGuard) *core.HTMLComponent {
	t.Helper()
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}
	shell := core.NewHTMLComponent("RevalShell", []byte(`<root>@include:outlet</root>`), nil)
	shell.SetComponent(shell)
	shell.AddDependency("outlet", NewOutlet())
	shell.Init(nil)

	RegisterRoute(Route{
		Path:         path,
		Component:    page,
		ResultGuards: []ResultGuard{guard},
		Loader:       func(context.Context, LoadContext) (any, error) { return "payload", nil },
		Meta:         map[string]any{"section": "session"},
	})
	MountRoot(shell)
	Navigate(path)

	dom.Query("[data-host-page='reval']").SetHTML(`<b id="reval-built">built</b>`)
	if !strings.Contains(dom.ByID("app").HTML(), "reval-built") {
		t.Fatal("the mounted page was not rendered")
	}
	return shell
}

// mustComponentRoot resolves a component's own root node. dom.ComponentRoot
// answers #app for an id it cannot resolve, so the id it carries is checked:
// asserting on the fallback would compare the shell to the container around it.
func mustComponentRoot(t *testing.T, id string) dom.Element {
	t.Helper()
	root := dom.ComponentRoot(id)
	if root.IsNull() || root.IsUndefined() || root.Attr("data-component-id") != id {
		t.Fatalf("component %s has no root of its own: %s", id, dom.ByID("app").HTML())
	}
	return root
}

// An allowed revalidation touches no DOM at all, and a refused one removes the
// routed subtree while the shell and the outlet marker around it stay the very
// nodes they were: the next navigation renders into the page the user is still
// looking at.
func TestRevalidateClearsTheRoutedSubtreeAndKeepsTheShell(t *testing.T) {
	Reset()
	restoreHistoryPath(t)

	allowed := true
	page := newHostGuardPage("RevalPage", "reval")
	shell := mountRevalidateShell(t, "/reval-dom", page, func(map[string]string) GuardResult {
		if !allowed {
			return Forbid()
		}
		return Allow()
	})
	if Data().Get() != "payload" || Meta().Get()["section"] != "session" {
		t.Fatalf("the route committed no state to revoke: data=%#v meta=%#v", Data().Get(), Meta().Get())
	}
	appNode := dom.ByID("app")
	shellNode := mustComponentRoot(t, shell.GetID())
	outletNode := dom.Query("[data-router-outlet]")
	historyLength := js.History().Get("length").Int()
	location := js.Location().Get("pathname").String()

	if err := RevalidateContext(context.Background()); err != nil {
		t.Fatalf("revalidate: %v", err)
	}
	if !strings.Contains(dom.ByID("app").HTML(), "reval-built") {
		t.Fatalf("an allowed revalidation re-rendered the page: %s", dom.ByID("app").HTML())
	}
	if page.unmounted != 0 {
		t.Fatalf("an allowed revalidation unmounted the page %d times", page.unmounted)
	}
	if got := mustComponentRoot(t, shell.GetID()); !got.Value.Equal(shellNode.Value) {
		t.Fatal("an allowed revalidation replaced the shell root")
	}

	allowed = false
	if err := RevalidateContext(context.Background()); !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden revalidation, got %v", err)
	}

	if page.unmounted != 1 {
		t.Fatalf("the page unmounted %d times, want 1", page.unmounted)
	}
	if CurrentComponent() != nil {
		t.Fatalf("the revoked route stayed current: %#v", CurrentComponent())
	}
	if root := dom.Query("[data-host-page='reval']"); !root.IsNull() {
		t.Fatalf("the protected subtree is still in the document: %s", dom.ByID("app").HTML())
	}
	if strings.Contains(dom.ByID("app").HTML(), "reval-built") {
		t.Fatalf("the protected DOM survived the revocation: %s", dom.ByID("app").HTML())
	}
	if got := dom.ByID("app"); !got.Value.Equal(appNode.Value) {
		t.Fatal("the revocation replaced the application container")
	}
	if got := mustComponentRoot(t, shell.GetID()); !got.Value.Equal(shellNode.Value) {
		t.Fatal("the revocation replaced the shell root MountRoot mounted")
	}
	if !strings.Contains(shellNode.HTML(), "data-router-outlet") {
		t.Fatalf("the revocation emptied the shell instead of the route: %s", shellNode.HTML())
	}
	after := dom.Query("[data-router-outlet]")
	if after.IsNull() || !after.Value.Equal(outletNode.Value) {
		t.Fatal("the revocation replaced the outlet marker instead of emptying it")
	}
	if html := after.HTML(); html != "" {
		t.Fatalf("the outlet still carries the routed subtree: %s", html)
	}
	if got := js.History().Get("length").Int(); got != historyLength {
		t.Fatalf("history length changed from %d to %d", historyLength, got)
	}
	if got := js.Location().Get("pathname").String(); got != location {
		t.Fatalf("the revocation moved the location to %q", got)
	}
	if got := ActivePath().Get(); got != "/reval-dom" {
		t.Fatalf("active path = %q, want the committed path", got)
	}
	if Status().Get() != NavigationError || !errors.Is(Error().Get(), ErrNavigationForbidden) {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
	// The DOM is not the only place the route left something behind: what it
	// loaded goes with the subtree it was rendered into.
	if Data().Get() != nil {
		t.Fatalf("the revoked route's data is still readable: %#v", Data().Get())
	}
	if meta := Meta().Get(); len(meta) != 0 {
		t.Fatalf("the revoked route's metadata is still readable: %#v", meta)
	}
}

// Without an outlet the router owns #app itself: the revocation removes the
// route's own root and leaves the container it was rendered into. #app is what
// dom.ComponentRoot falls back to for an id it cannot resolve, and that
// fallback must never be cleared in the route's place.
func TestRevalidateWithoutAnOutletClearsOnlyTheRouteRoot(t *testing.T) {
	Reset()
	restoreHistoryPath(t)
	previousOutlet := liveOutlet
	liveOutlet = nil
	t.Cleanup(func() { liveOutlet = previousOutlet })

	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}
	allowed := true
	page := newHostGuardPage("RevalBarePage", "bare")
	RegisterRoute(Route{
		Path:      "/reval-bare",
		Component: page,
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			if !allowed {
				return Forbid()
			}
			return Allow()
		}},
	})
	Navigate("/reval-bare")
	if dom.Query("[data-host-page='bare']").IsNull() {
		t.Fatalf("the page was not rendered: %s", dom.ByID("app").HTML())
	}
	appNode := dom.ByID("app")

	allowed = false
	if err := RevalidateContext(context.Background()); !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden revalidation, got %v", err)
	}

	if page.unmounted != 1 {
		t.Fatalf("the page unmounted %d times, want 1", page.unmounted)
	}
	if root := dom.Query("[data-host-page='bare']"); !root.IsNull() {
		t.Fatalf("the protected subtree is still in the document: %s", dom.ByID("app").HTML())
	}
	if got := dom.ByID("app"); got.IsNull() || !got.Value.Equal(appNode.Value) {
		t.Fatal("the revocation removed the application container instead of the route")
	}
}

// A guard that hands off during revalidation performs the browser navigation
// and commits nothing on the way out: the page the user is on survives
// untouched until the browser unloads it.
func TestRevalidateHostReplaceHandsOffWithoutTearingDown(t *testing.T) {
	Reset()
	restoreHistoryPath(t)
	calls := captureHostReplace(t)

	valid := true
	page := newHostGuardPage("RevalHostPage", "reval")
	mountRevalidateShell(t, "/reval-host-dom", page, func(map[string]string) GuardResult {
		if !valid {
			return HostReplace("/login?next=%2Freval-host-dom")
		}
		return Allow()
	})
	historyLength := js.History().Get("length").Int()
	location := js.Location().Get("pathname").String()

	valid = false
	if err := RevalidateContext(context.Background()); err != nil {
		t.Fatalf("host handoff: %v", err)
	}

	if len(*calls) != 1 || (*calls)[0] != "/login?next=%2Freval-host-dom" {
		t.Fatalf("host replace calls = %#v, want one handoff", *calls)
	}
	if page.unmounted != 0 {
		t.Fatalf("the mounted page was unmounted %d times before the handoff", page.unmounted)
	}
	if CurrentComponent() != core.Component(page) {
		t.Fatalf("current component changed to %#v", CurrentComponent())
	}
	if !strings.Contains(dom.ByID("app").HTML(), "reval-built") {
		t.Fatalf("the handoff tore down the mounted page: %s", dom.ByID("app").HTML())
	}
	if got := js.History().Get("length").Int(); got != historyLength {
		t.Fatalf("history length changed from %d to %d", historyLength, got)
	}
	if got := js.Location().Get("pathname").String(); got != location {
		t.Fatalf("the SPA moved the location to %q", got)
	}
	if got := ActivePath().Get(); got != "/reval-host-dom" {
		t.Fatalf("active path = %q", got)
	}
	if Status().Get() != NavigationReady || Error().Get() != nil {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
	// Nothing was taken back, so the page still holds what it loaded until the
	// browser unloads it.
	if Data().Get() != "payload" || Meta().Get()["section"] != "session" {
		t.Fatalf("the handoff changed route state: data=%#v meta=%#v", Data().Get(), Meta().Get())
	}
}

// Both redirect actions replace the entry the guard refused: back returns to
// the page before the protected one, never to the protected one itself.
func TestRevalidateRedirectReplacesTheRefusedEntry(t *testing.T) {
	for name, decision := range map[string]GuardResult{
		"redirect": RedirectTo("/reval-hist-login"),
		"replace":  ReplaceWith("/reval-hist-login"),
	} {
		t.Run(name, func(t *testing.T) {
			Reset()
			restoreHistoryPath(t)
			result := decision
			allowed := true

			RegisterRoute(Route{Path: "/reval-hist-base", Component: func() core.Component { return routeComponent{} }})
			RegisterRoute(Route{Path: "/reval-hist-login", Component: func() core.Component { return routeComponent{} }})
			page := newHostGuardPage("RevalHistPage", "hist")
			RegisterRoute(Route{
				Path:      "/reval-hist-account",
				Component: page,
				ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
					if !allowed {
						return result
					}
					return Allow()
				}},
			})

			Navigate("/reval-hist-base")
			Navigate("/reval-hist-account")
			before := js.History().Get("length").Int()

			allowed = false
			if err := RevalidateContext(context.Background()); err != nil {
				t.Fatalf("revalidate: %v", err)
			}

			if got := js.Location().Get("pathname").String(); got != "/reval-hist-login" {
				t.Fatalf("path after the redirect = %q", got)
			}
			if got := js.History().Get("length").Int(); got != before {
				t.Fatalf("history length changed from %d to %d", before, got)
			}
			if page.unmounted != 1 {
				t.Fatalf("the refused page unmounted %d times, want 1", page.unmounted)
			}
			if got := ActivePath().Get(); got != "/reval-hist-login" {
				t.Fatalf("active path = %q", got)
			}
			if got := travelHistory(t, "back"); got != "/reval-hist-base" {
				t.Fatalf("back landed on %q, want /reval-hist-base", got)
			}
		})
	}
}

// The revocation leaves the refused path in the address bar, so the browser can
// come back to it. It does, through the guards: a back and a forward cannot put
// the protected page back on screen.
func TestRevalidateRevocationSurvivesBackAndForward(t *testing.T) {
	Reset()
	restoreHistoryPath(t)

	allowed, guards := true, 0
	RegisterRoute(Route{Path: "/reval-back-base", Component: func() core.Component { return routeComponent{} }})
	page := newHostGuardPage("RevalBackPage", "back")
	RegisterRoute(Route{
		Path:      "/reval-back-guarded",
		Component: page,
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			guards++
			if !allowed {
				return Forbid()
			}
			return Allow()
		}},
	})

	Navigate("/reval-back-base")
	Navigate("/reval-back-guarded")

	allowed = false
	if err := RevalidateContext(context.Background()); !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden revalidation, got %v", err)
	}
	if root := dom.Query("[data-host-page='back']"); !root.IsNull() {
		t.Fatal("the protected subtree survived the revocation")
	}

	back := travelHistory(t, "back")
	if back != "/reval-back-base" {
		t.Fatalf("back landed on %q, want /reval-back-base", back)
	}
	// What the popstate listener InitRouter installs does with the new URL.
	navigate(back, historyNone)
	before := guards

	forward := travelHistory(t, "forward")
	if forward != "/reval-back-guarded" {
		t.Fatalf("forward landed on %q, want /reval-back-guarded", forward)
	}
	navigate(forward, historyNone)

	if guards <= before {
		t.Fatal("forward restored the refused path without running its guards")
	}
	if root := dom.Query("[data-host-page='back']"); !root.IsNull() {
		t.Fatalf("forward put the protected page back on screen: %s", dom.ByID("app").HTML())
	}
	if CurrentComponent() == core.Component(page) {
		t.Fatal("forward remounted the revoked page")
	}
	if page.unmounted != 1 {
		t.Fatalf("the page unmounted %d times, want 1", page.unmounted)
	}
}

// revalidateGate stands in for the ownership gate a host registration hands to
// the frames delivered under it: releasing the registration closes it.
type revalidateGate struct{ closed bool }

func (g *revalidateGate) open() bool { return !g.closed }

// The revocation runs the component's own cleanup: the scope closes (the
// disconnect a long-lived session such as a noVNC client hangs off it), the
// owned host registration is released, and the signals go with it. A frame
// still in flight for that component then finds no root, no signals, and a
// closed gate.
func TestRevalidateReleasesOwnedHostWorkAndIgnoresLateFrames(t *testing.T) {
	Reset()
	restoreHistoryPath(t)

	gate := &revalidateGate{}
	releases := 0
	core.SetHostRegistrar(func(string, string, []string) func() {
		return func() {
			releases++
			gate.closed = true
		}
	})
	t.Cleanup(func() { core.SetHostRegistrar(nil) })

	allowed := true
	page := newHostGuardPage("RevalOwnedPage", "reval")
	// Declared before the mount, the way a page that talks to a host component
	// declares it in its constructor: the mount is what opens the registration.
	page.AddHostComponent("RevalHostComponent")
	mountRevalidateShell(t, "/reval-host-owned", page, func(map[string]string) GuardResult {
		if !allowed {
			return Forbid()
		}
		return Allow()
	})

	disconnects := 0
	page.Scope().Defer(func() { disconnects++ })
	signal := state.NewSignal("connected")
	dom.RegisterSignal(page.GetID(), "status", signal)

	if releases != 0 {
		t.Fatalf("the registration was released %d times before the revocation", releases)
	}

	allowed = false
	if err := RevalidateContext(context.Background()); !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden revalidation, got %v", err)
	}

	if releases != 1 {
		t.Fatalf("the host registration was released %d times, want 1", releases)
	}
	if disconnects != 1 {
		t.Fatalf("scope cleanup ran %d times, want 1", disconnects)
	}
	if signals := dom.SnapshotComponentSignals(page.GetID()); len(signals) != 0 {
		t.Fatalf("the revoked component still exposes host signals: %#v", signals)
	}
	// A late frame resolves its target the way host delivery does, on the exact
	// id and never through the #app fallback.
	if root := dom.Doc().Query("[data-component-id='" + page.GetID() + "']"); !root.IsNull() {
		t.Fatal("a late frame would still find the revoked component's root")
	}
	if signal.SetFromHostGated("hijacked", gate.open) {
		t.Fatal("a late host frame wrote into the revoked component's signal")
	}
	if got := signal.Get(); got != "connected" {
		t.Fatalf("the revoked component's signal changed to %q", got)
	}
}
