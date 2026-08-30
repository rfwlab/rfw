package router

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
)

// guardedRoute registers a protected route whose loader and component factory
// must never run when a guard refuses the destination.
func guardedRoute(t *testing.T, path string, guards []ResultGuard, legacy ...Guard) *int {
	t.Helper()
	created := 0
	RegisterRoute(Route{
		Path:         path,
		Guards:       legacy,
		ResultGuards: guards,
		Component: func() core.Component {
			created++
			return &recordComponent{name: "protected"}
		},
		Loader: func(context.Context, LoadContext) (any, error) {
			t.Error("route loader ran for a guarded destination")
			return nil, nil
		},
	})
	return &created
}

func TestResultGuardForbidsWithoutCommitting(t *testing.T) {
	resetRouter(t)
	created := guardedRoute(t, "/vault", []ResultGuard{func(map[string]string) GuardResult {
		return Forbid()
	}})

	err := NavigateContext(context.Background(), "/vault")

	if !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden navigation, got %v", err)
	}
	if *created != 0 {
		t.Fatalf("protected component was created %d times", *created)
	}
	if CurrentComponent() != nil {
		t.Fatalf("forbidden navigation committed %#v", CurrentComponent())
	}
	if Status().Get() != NavigationError || !errors.Is(Error().Get(), ErrNavigationForbidden) {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
}

// Guards run parent before child, legacy before typed, and stop at the first
// one that does not allow navigation.
func TestResultGuardsComposeParentBeforeChild(t *testing.T) {
	resetRouter(t)
	var order []string
	record := func(name string) { order = append(order, name) }
	RegisterRoute(Route{
		Path:   "/app",
		Guards: []Guard{func(map[string]string) bool { record("parent-legacy"); return true }},
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			record("parent-result")
			return Allow()
		}},
		Children: []Route{{
			Path:      "settings",
			Guards:    []Guard{func(map[string]string) bool { record("child-legacy"); return true }},
			Component: func() core.Component { return &recordComponent{name: "settings"} },
			ResultGuards: []ResultGuard{
				func(map[string]string) GuardResult { record("child-result"); return Forbid() },
				func(map[string]string) GuardResult { record("unreachable"); return Allow() },
			},
		}},
	})

	if err := NavigateContext(context.Background(), "/app/settings"); !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden navigation, got %v", err)
	}

	want := []string{"parent-legacy", "parent-result", "child-legacy", "child-result"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("guard order = %v, want %v", order, want)
	}
}

// A parent guard that refuses stops the chain before any child guard runs.
func TestParentResultGuardStopsChildGuards(t *testing.T) {
	resetRouter(t)
	childRan := false
	RegisterRoute(Route{
		Path: "/admin",
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			return ReplaceWith("/login")
		}},
		Children: []Route{{
			Path:      "users",
			Component: func() core.Component { return &recordComponent{name: "users"} },
			ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
				childRan = true
				return Allow()
			}},
		}},
	})
	RegisterRoute(Route{Path: "/login", Component: func() core.Component { return &recordComponent{name: "login"} }})

	if err := NavigateContext(context.Background(), "/admin/users"); err != nil {
		t.Fatalf("guard redirect: %v", err)
	}
	if childRan {
		t.Fatal("a child guard ran after the parent refused")
	}
	if got := mustRecord(t, CurrentComponent()).name; got != "login" {
		t.Fatalf("current component = %q, want login", got)
	}
}

func TestResultGuardRedirectsToLogin(t *testing.T) {
	resetRouter(t)
	RegisterRoute(Route{Path: "/login", Component: func() core.Component { return &recordComponent{name: "login"} }})
	created := guardedRoute(t, "/account", []ResultGuard{func(map[string]string) GuardResult {
		return ReplaceWith("/login")
	}})

	if err := NavigateContext(context.Background(), "/account"); err != nil {
		t.Fatalf("guard redirect: %v", err)
	}
	if *created != 0 {
		t.Fatalf("protected component was created %d times", *created)
	}
	if got := mustRecord(t, CurrentComponent()).name; got != "login" {
		t.Fatalf("current component = %q, want login", got)
	}
	if ActivePath().Get() != "/login" {
		t.Fatalf("active path = %q, want /login", ActivePath().Get())
	}
	if Status().Get() != NavigationReady {
		t.Fatalf("navigation status = %s, want ready", Status().Get())
	}
}

func TestResultGuardRedirectKeepsRouteParameters(t *testing.T) {
	resetRouter(t)
	var seen map[string]string
	RegisterRoute(Route{Path: "/denied", Component: func() core.Component { return &recordComponent{name: "denied"} }})
	RegisterRoute(Route{
		Path:      "/orders/:id",
		Component: func() core.Component { return &recordComponent{name: "order"} },
		ResultGuards: []ResultGuard{func(params map[string]string) GuardResult {
			seen = params
			return RedirectTo("/denied")
		}},
	})

	if err := NavigateContext(context.Background(), "/orders/17?ref=mail"); err != nil {
		t.Fatalf("guard redirect: %v", err)
	}
	if seen["id"] != "17" || seen["ref"] != "mail" {
		t.Fatalf("guard parameters = %#v", seen)
	}
	if got := mustRecord(t, CurrentComponent()).name; got != "denied" {
		t.Fatalf("current component = %q, want denied", got)
	}
}

func TestInvalidGuardResultFailsClosed(t *testing.T) {
	cases := map[string]GuardResult{
		"empty":           RedirectTo(""),
		"blank":           ReplaceWith("   "),
		"relative":        RedirectTo("login"),
		"scheme relative": RedirectTo("//evil.example/login"),
		"unknown action":  {Action: GuardAction(42), Path: "/login"},
	}
	for name, result := range cases {
		t.Run(name, func(t *testing.T) {
			resetRouter(t)
			decision := result
			created := guardedRoute(t, "/secret", []ResultGuard{func(map[string]string) GuardResult {
				return decision
			}})

			err := NavigateContext(context.Background(), "/secret")

			if !errors.Is(err, ErrInvalidGuardResult) {
				t.Fatalf("expected an invalid guard result, got %v", err)
			}
			if *created != 0 || CurrentComponent() != nil {
				t.Fatalf("invalid guard result committed a route: created=%d component=%#v", *created, CurrentComponent())
			}
		})
	}
}

func TestResultGuardRedirectLoopFails(t *testing.T) {
	resetRouter(t)
	RegisterRoute(Route{
		Path:         "/ping",
		Component:    func() core.Component { return &recordComponent{name: "ping"} },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return RedirectTo("/pong") }},
	})
	RegisterRoute(Route{
		Path:         "/pong",
		Component:    func() core.Component { return &recordComponent{name: "pong"} },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return ReplaceWith("/ping") }},
	})

	if err := NavigateContext(context.Background(), "/ping"); !errors.Is(err, ErrRedirectLoop) {
		t.Fatalf("expected a redirect loop error, got %v", err)
	}
	if CurrentComponent() != nil {
		t.Fatalf("redirect loop committed %#v", CurrentComponent())
	}
}

// Legacy bool guards keep their contract: a false guard blocks with
// ErrNavigationBlocked and never reaches the typed guards behind it.
func TestLegacyGuardStillBlocksBeforeResultGuards(t *testing.T) {
	resetRouter(t)
	resultRan := false
	RegisterRoute(Route{Path: "/", Component: func() core.Component { return &recordComponent{name: "root"} }})
	RegisterRoute(Route{
		Path:      "/legacy",
		Component: func() core.Component { return &recordComponent{name: "legacy"} },
		Guards:    []Guard{func(map[string]string) bool { return false }},
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			resultRan = true
			return Allow()
		}},
	})

	if err := NavigateContext(context.Background(), "/legacy"); !errors.Is(err, ErrNavigationBlocked) {
		t.Fatalf("expected a blocked navigation, got %v", err)
	}
	if resultRan {
		t.Fatal("a result guard ran after a legacy guard blocked")
	}
}

// A guard that allows leaves the route untouched: loader, component and route
// state commit exactly as they do without guards.
func TestAllowingResultGuardCommitsRoute(t *testing.T) {
	resetRouter(t)
	loaded := false
	RegisterRoute(Route{
		Path:         "/open",
		Component:    func() core.Component { return &recordComponent{name: "open"} },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return Allow() }},
		Loader: func(context.Context, LoadContext) (any, error) {
			loaded = true
			return nil, nil
		},
	})

	if err := NavigateContext(context.Background(), "/open"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	if !loaded {
		t.Fatal("an allowed route did not run its loader")
	}
	if got := mustRecord(t, CurrentComponent()).name; got != "open" {
		t.Fatalf("current component = %q, want open", got)
	}
}

// A cancelled navigation never reaches its guards.
func TestCancelledNavigationSkipsResultGuards(t *testing.T) {
	resetRouter(t)
	guardRan := false
	RegisterRoute(Route{
		Path:      "/cancelled-guard",
		Component: func() core.Component { return &recordComponent{name: "cancelled"} },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			guardRan = true
			return Allow()
		}},
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := NavigateContext(ctx, "/cancelled-guard"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected a cancelled navigation, got %v", err)
	}
	if guardRan {
		t.Fatal("a guard ran for a cancelled navigation")
	}
}
