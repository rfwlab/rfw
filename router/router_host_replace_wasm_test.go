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
)

// hostGuardPage is a routed page that counts its unmounts, so a test can tell
// a page left in place from one the router tore down.
type hostGuardPage struct {
	*core.HTMLComponent
	unmounted int
}

func (p *hostGuardPage) Unmount() {
	p.unmounted++
	p.HTMLComponent.Unmount()
}

func newHostGuardPage(name, marker string) *hostGuardPage {
	page := &hostGuardPage{HTMLComponent: core.NewHTMLComponent(
		name,
		[]byte(`<root><div data-host-page="`+marker+`"></div></root>`),
		nil,
	)}
	page.SetComponent(page)
	page.Init(nil)
	return page
}

// captureHostReplace swaps the browser handoff for a recorder. The production
// path calls location.replace, which would unload the page running the test:
// this proves the router asked for the handoff and with which path, not that
// the host served the document.
func captureHostReplace(t *testing.T) *[]string {
	t.Helper()
	previous := hostReplace
	calls := []string{}
	hostReplace = func(path string) { calls = append(calls, path) }
	t.Cleanup(func() { hostReplace = previous })
	return &calls
}

// mountHostShell renders a shell with a router outlet into #app and returns the
// page mounted on path, with a marker injected after mount the way a fetch
// callback fills a card. A remount would wipe that marker.
func mountHostShell(t *testing.T, path string) *hostGuardPage {
	t.Helper()
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}
	shell := core.NewHTMLComponent("HostShell", []byte(`<root>@include:outlet</root>`), nil)
	shell.SetComponent(shell)
	shell.AddDependency("outlet", NewOutlet())
	shell.Init(nil)

	page := newHostGuardPage("HostDashboard", "dashboard")
	Page(path, page)
	MountRoot(shell)
	Navigate(path)

	dom.Query("[data-host-page='dashboard']").SetHTML(`<b id="host-built">built</b>`)
	if !strings.Contains(dom.ByID("app").HTML(), "host-built") {
		t.Fatal("the mounted page was not rendered")
	}
	return page
}

// A guard that hands the document to a host-owned path performs the browser
// navigation and commits nothing on the way out: the protected route's loader,
// component factory, DOM and route state stay untouched, no history entry is
// written, and the page the user is looking at survives unchanged until the
// browser unloads it.
func TestHostReplaceHandsOffWithoutCommitting(t *testing.T) {
	Reset()
	restoreHistoryPath(t)
	calls := captureHostReplace(t)

	dashboard := mountHostShell(t, "/host-dashboard")
	created := 0
	RegisterRoute(Route{
		Path:         "/host-admin",
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return HostReplace("/login") }},
		Component: func() core.Component {
			created++
			return newHostGuardPage("HostAdmin", "admin")
		},
		Loader: func(context.Context, LoadContext) (any, error) {
			t.Error("the protected route ran its loader before the host handoff")
			return nil, nil
		},
		Meta: map[string]any{"section": "admin"},
	})
	NotFoundCallback = func(string) { t.Error("the host handoff went through the not-found callback") }
	NotFoundComponent = func() core.Component {
		t.Error("the host handoff rendered the not-found component")
		return newHostGuardPage("HostNotFound", "not-found")
	}

	historyLength := js.History().Get("length").Int()
	location := js.Location().Get("pathname").String()

	if err := NavigateContext(context.Background(), "/host-admin"); err != nil {
		t.Fatalf("host handoff: %v", err)
	}

	if len(*calls) != 1 || (*calls)[0] != "/login" {
		t.Fatalf("host replace calls = %#v, want one call with /login", *calls)
	}
	if created != 0 {
		t.Fatalf("the protected component was created %d times", created)
	}
	if got := js.History().Get("length").Int(); got != historyLength {
		t.Fatalf("history length changed from %d to %d", historyLength, got)
	}
	if got := js.Location().Get("pathname").String(); got != location {
		t.Fatalf("the SPA moved the location to %q, want %q", got, location)
	}
	if got := ActivePath().Get(); got != "/host-dashboard" {
		t.Fatalf("active path = %q, want /host-dashboard", got)
	}
	if mounted, ok := CurrentComponent().(*hostGuardPage); !ok || mounted != dashboard {
		t.Fatalf("current component changed to %#v", CurrentComponent())
	}
	if dashboard.unmounted != 0 {
		t.Fatalf("the mounted page was unmounted %d times before the handoff", dashboard.unmounted)
	}
	html := dom.ByID("app").HTML()
	if !strings.Contains(html, "host-built") {
		t.Fatalf("the handoff tore down the mounted page: %s", html)
	}
	if strings.Contains(html, `data-host-page="admin"`) {
		t.Fatalf("the protected page reached the DOM: %s", html)
	}
	if Status().Get() != NavigationReady {
		t.Fatalf("navigation status = %s, want the mounted route's ready", Status().Get())
	}
	if Error().Get() != nil {
		t.Fatalf("navigation error = %v, want none", Error().Get())
	}
	if section, ok := Meta().Get()["section"]; ok {
		t.Fatalf("the protected route committed its metadata: %v", section)
	}
}

// A parent guard that hands off stops there: no child guard, no child loader,
// no component.
func TestParentHostReplaceStopsChildGuards(t *testing.T) {
	Reset()
	restoreHistoryPath(t)
	calls := captureHostReplace(t)

	childRan, created := false, 0
	RegisterRoute(Route{
		Path:         "/host-parent",
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return HostReplace("/login") }},
		Children: []Route{{
			Path: "users",
			ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
				childRan = true
				return Allow()
			}},
			Component: func() core.Component {
				created++
				return routeComponent{}
			},
			Loader: func(context.Context, LoadContext) (any, error) {
				t.Error("a child loader ran after the parent handed off")
				return nil, nil
			},
		}},
	})

	if err := NavigateContext(context.Background(), "/host-parent/users"); err != nil {
		t.Fatalf("host handoff: %v", err)
	}

	if len(*calls) != 1 || (*calls)[0] != "/login" {
		t.Fatalf("host replace calls = %#v, want one call with /login", *calls)
	}
	if childRan {
		t.Fatal("a child guard ran after the parent handed off")
	}
	if created != 0 {
		t.Fatalf("the child component was created %d times", created)
	}
}

// An unauthenticated caller leaves for the host login page while the same guard
// keeps a permission refusal inside the SPA.
func TestSessionGuardHandsOffUnauthenticated(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	calls := captureHostReplace(t)

	authenticated, permitted := false, true
	created := sessionGuardedRoute(t, &authenticated, &permitted)

	if err := NavigateContext(context.Background(), "/composed"); err != nil {
		t.Fatalf("host handoff: %v", err)
	}

	if len(*calls) != 1 || (*calls)[0] != "/login" {
		t.Fatalf("host replace calls = %#v, want one call with /login", *calls)
	}
	if *created != 0 || CurrentComponent() != nil {
		t.Fatalf("the handoff committed a route: created=%d component=%#v", *created, CurrentComponent())
	}
}

// The accepted targets reach the handoff exactly as validation returned them.
func TestHostReplaceHandsOffAcceptedTargets(t *testing.T) {
	for _, target := range hostReplaceAccepts {
		t.Run(target, func(t *testing.T) {
			Reset()
			t.Cleanup(Reset)
			calls := captureHostReplace(t)
			destination := target
			created := guardedRoute(t, "/host-accepted", []ResultGuard{func(map[string]string) GuardResult {
				return HostReplace(destination)
			}})

			if err := NavigateContext(context.Background(), "/host-accepted"); err != nil {
				t.Fatalf("host handoff: %v", err)
			}

			if len(*calls) != 1 || (*calls)[0] != target {
				t.Fatalf("host replace calls = %#v, want one call with %q", *calls, target)
			}
			if *created != 0 {
				t.Fatalf("the protected component was created %d times", *created)
			}
		})
	}
}

// A target validation refuses never reaches the browser: navigation fails
// closed and the document stays where it is.
func TestHostReplaceRejectedTargetsNeverHandOff(t *testing.T) {
	for _, tc := range hostReplaceRejects {
		t.Run(tc.name, func(t *testing.T) {
			Reset()
			restoreHistoryPath(t)
			calls := captureHostReplace(t)
			destination := tc.target
			created := guardedRoute(t, "/host-rejected", []ResultGuard{func(map[string]string) GuardResult {
				return HostReplace(destination)
			}})
			location := js.Location().Get("pathname").String()

			err := NavigateContext(context.Background(), "/host-rejected")

			if !errors.Is(err, ErrInvalidGuardResult) {
				t.Fatalf("expected an invalid guard result, got %v", err)
			}
			if len(*calls) != 0 {
				t.Fatalf("a rejected target reached the browser: %#v", *calls)
			}
			if *created != 0 || CurrentComponent() != nil {
				t.Fatalf("a rejected target committed a route: created=%d component=%#v", *created, CurrentComponent())
			}
			if got := js.Location().Get("pathname").String(); got != location {
				t.Fatalf("location moved to %q, want %q", got, location)
			}
		})
	}
}

// The SPA redirects are untouched by the new action: they keep matching routes
// and writing history entries, and never reach the host handoff.
func TestGuardRedirectsStayInsideTheSPA(t *testing.T) {
	Reset()
	restoreHistoryPath(t)
	calls := captureHostReplace(t)

	RegisterRoute(Route{Path: "/spa-origin", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{Path: "/spa-offers", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{Path: "/spa-login", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{
		Path:         "/spa-promo",
		Component:    func() core.Component { return routeComponent{} },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return RedirectTo("/spa-offers") }},
	})
	RegisterRoute(Route{
		Path:         "/spa-account",
		Component:    func() core.Component { return routeComponent{} },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return ReplaceWith("/spa-login") }},
	})

	Navigate("/spa-origin")
	before := js.History().Get("length").Int()

	Navigate("/spa-promo")

	if got := js.Location().Get("pathname").String(); got != "/spa-offers" {
		t.Fatalf("path after the guard redirect = %q", got)
	}
	if got := js.History().Get("length").Int(); got != before+1 {
		t.Fatalf("history length = %d, want %d", got, before+1)
	}

	before = js.History().Get("length").Int()

	Navigate("/spa-account")

	if got := js.Location().Get("pathname").String(); got != "/spa-login" {
		t.Fatalf("path after the guard replace = %q", got)
	}
	if got := js.History().Get("length").Int(); got != before {
		t.Fatalf("history length changed from %d to %d", before, got)
	}
	if len(*calls) != 0 {
		t.Fatalf("an SPA redirect reached the host handoff: %#v", *calls)
	}
}
