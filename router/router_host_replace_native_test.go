//go:build !js || !wasm

package router

import (
	"context"
	"errors"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
)

// hostNavigation asserts that err is the host handoff contract and returns the
// path it carries.
func hostNavigation(t *testing.T, err error) string {
	t.Helper()
	if !errors.Is(err, ErrHostNavigation) {
		t.Fatalf("expected a host navigation, got %v", err)
	}
	var host *HostNavigationError
	if !errors.As(err, &host) {
		t.Fatalf("host navigation error does not carry its target: %v", err)
	}
	return host.Path
}

// Outside the browser there is no document to replace, so the router reports
// the handoff with its target instead of pretending the host page loaded.
func TestHostReplaceReportsHandoffWithoutCommitting(t *testing.T) {
	resetRouter(t)
	RegisterRoute(Route{Path: "/", Component: func() core.Component { return &recordComponent{name: "root"} }})
	Navigate("/")
	created := 0
	RegisterRoute(Route{
		Path: "/admin",
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			return HostReplace("/login?next=%2Fadmin")
		}},
		Component: func() core.Component {
			created++
			return &recordComponent{name: "admin"}
		},
		Loader: func(context.Context, LoadContext) (any, error) {
			t.Error("the protected route ran its loader before the host handoff")
			return nil, nil
		},
		Meta: map[string]any{"section": "admin"},
	})

	err := NavigateContext(context.Background(), "/admin")

	if got := hostNavigation(t, err); got != "/login?next=%2Fadmin" {
		t.Fatalf("host path = %q, want /login?next=%%2Fadmin", got)
	}
	if created != 0 {
		t.Fatalf("the protected component was created %d times", created)
	}
	if got := mustRecord(t, CurrentComponent()).name; got != "root" {
		t.Fatalf("current component = %q, want the mounted root", got)
	}
	if got := ActivePath().Get(); got != "/" {
		t.Fatalf("active path = %q, want /", got)
	}
	if Data().Get() != nil {
		t.Fatalf("the handoff committed route data: %#v", Data().Get())
	}
	if section, ok := Meta().Get()["section"]; ok {
		t.Fatalf("the handoff committed route metadata: %v", section)
	}
	if Status().Get() != NavigationError || !errors.Is(Error().Get(), ErrHostNavigation) {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
}

func TestHostReplaceReportsAcceptedTargets(t *testing.T) {
	for _, target := range hostReplaceAccepts {
		t.Run(target, func(t *testing.T) {
			resetRouter(t)
			destination := target
			created := guardedRoute(t, "/host-accepted", []ResultGuard{func(map[string]string) GuardResult {
				return HostReplace(destination)
			}})

			err := NavigateContext(context.Background(), "/host-accepted")

			if got := hostNavigation(t, err); got != target {
				t.Fatalf("host path = %q, want %q", got, target)
			}
			if *created != 0 || CurrentComponent() != nil {
				t.Fatalf("the handoff committed a route: created=%d component=%#v", *created, CurrentComponent())
			}
		})
	}
}

// A target validation refuses fails closed with ErrInvalidGuardResult and never
// becomes a host handoff.
func TestHostReplaceRejectedTargetsFailClosed(t *testing.T) {
	for _, tc := range hostReplaceRejects {
		t.Run(tc.name, func(t *testing.T) {
			resetRouter(t)
			destination := tc.target
			created := guardedRoute(t, "/host-rejected", []ResultGuard{func(map[string]string) GuardResult {
				return HostReplace(destination)
			}})

			err := NavigateContext(context.Background(), "/host-rejected")

			if !errors.Is(err, ErrInvalidGuardResult) {
				t.Fatalf("expected an invalid guard result, got %v", err)
			}
			if errors.Is(err, ErrHostNavigation) {
				t.Fatalf("a rejected target became a host handoff: %v", err)
			}
			if *created != 0 || CurrentComponent() != nil {
				t.Fatalf("a rejected target committed a route: created=%d component=%#v", *created, CurrentComponent())
			}
		})
	}
}

// A parent guard that hands off stops the chain: no child guard, no child
// loader, no component.
func TestParentHostReplaceStopsChildGuardsOnStub(t *testing.T) {
	resetRouter(t)
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
				return &recordComponent{name: "users"}
			},
			Loader: func(context.Context, LoadContext) (any, error) {
				t.Error("a child loader ran after the parent handed off")
				return nil, nil
			},
		}},
	})

	err := NavigateContext(context.Background(), "/host-parent/users")

	if got := hostNavigation(t, err); got != "/login" {
		t.Fatalf("host path = %q, want /login", got)
	}
	if childRan {
		t.Fatal("a child guard ran after the parent handed off")
	}
	if created != 0 {
		t.Fatalf("the child component was created %d times", created)
	}
}

// An unauthenticated caller is handed to the host login page by the same guard
// that keeps a permission refusal inside the SPA.
func TestSessionGuardReportsHandoffForUnauthenticated(t *testing.T) {
	resetRouter(t)
	authenticated, permitted := false, true
	created := sessionGuardedRoute(t, &authenticated, &permitted)

	err := NavigateContext(context.Background(), "/composed")

	if got := hostNavigation(t, err); got != "/login" {
		t.Fatalf("host path = %q, want /login", got)
	}
	if *created != 0 || CurrentComponent() != nil {
		t.Fatalf("the handoff committed a route: created=%d component=%#v", *created, CurrentComponent())
	}
}
