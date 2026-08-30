//go:build !js || !wasm

package router

import (
	"context"
	"errors"
	"testing"
)

// Outside the browser there is no document to hand to the host, so a
// revalidation that hands off reports the target and leaves the mounted route
// exactly where it is, the contract NavigateContext already keeps.
func TestRevalidateHostReplaceReportsHandoff(t *testing.T) {
	resetRouter(t)
	valid := true
	session := registerSessionRoute(t, "/reval-host", func(map[string]string) GuardResult {
		if !valid {
			return HostReplace("/login?next=%2Freval-host")
		}
		return Allow()
	})

	if err := NavigateContext(context.Background(), "/reval-host"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())
	created, loaded := session.created, session.loaded

	valid = false
	err := RevalidateContext(context.Background())

	if got := hostNavigation(t, err); got != "/login?next=%2Freval-host" {
		t.Fatalf("host path = %q", got)
	}
	if page.unmounted != 0 || page.cleanups != 0 {
		t.Fatalf("the handoff tore the mounted page down: unmounted=%d cleanups=%d", page.unmounted, page.cleanups)
	}
	if CurrentComponent() != page {
		t.Fatalf("current component changed to %#v", CurrentComponent())
	}
	if session.created != created || session.loaded != loaded {
		t.Fatalf("the handoff re-ran the route: created=%d loaded=%d", session.created, session.loaded)
	}
	if got := ActivePath().Get(); got != "/reval-host" {
		t.Fatalf("active path = %q", got)
	}
	if Data().Get() != "payload" || Meta().Get()["section"] != "session" {
		t.Fatalf("the handoff changed route state: data=%#v meta=%#v", Data().Get(), Meta().Get())
	}
	if Status().Get() != NavigationError || !errors.Is(Error().Get(), ErrHostNavigation) {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
}

// A guard result validation refuses is a refusal like any other: the route the
// user is on is mounted, so it fails closed and is torn down rather than left
// in place on a decision the router could not apply.
func TestRevalidateInvalidGuardResultFailsClosed(t *testing.T) {
	for name, decision := range map[string]GuardResult{
		"scheme relative": RedirectTo("//evil.example/login"),
		"relative":        ReplaceWith("login"),
		"off origin host": HostReplace("https://evil.example/login"),
		"unknown action":  {Action: GuardAction(42), Path: "/login"},
	} {
		t.Run(name, func(t *testing.T) {
			resetRouter(t)
			result := decision
			valid := true
			registerSessionRoute(t, "/reval-invalid", func(map[string]string) GuardResult {
				if !valid {
					return result
				}
				return Allow()
			})

			if err := NavigateContext(context.Background(), "/reval-invalid"); err != nil {
				t.Fatalf("navigate: %v", err)
			}
			page := mustRevalidatePage(t, CurrentComponent())

			valid = false
			if err := RevalidateContext(context.Background()); !errors.Is(err, ErrInvalidGuardResult) {
				t.Fatalf("expected an invalid guard result, got %v", err)
			}
			if page.unmounted != 1 || page.cleanups != 1 {
				t.Fatalf("teardown ran %d times (cleanups %d), want 1", page.unmounted, page.cleanups)
			}
			if CurrentComponent() != nil {
				t.Fatalf("the route stayed current: %#v", CurrentComponent())
			}
			// A refusal the router could not even apply revokes like any other.
			if Data().Get() != nil || len(Meta().Get()) != 0 {
				t.Fatalf("the revoked route's state survived: data=%#v meta=%#v", Data().Get(), Meta().Get())
			}
		})
	}
}
