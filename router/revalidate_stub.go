//go:build !js || !wasm

package router

import (
	"context"

	"github.com/rfwlab/rfw/v2/core"
)

// routeUnmounter is a routed component that owns lifecycle cleanup. The
// Component interface outside browser builds is render-only, so a revoked route
// is unmounted through this one when the component provides it.
type routeUnmounter interface{ Unmount() }

// Revalidate re-runs the guards of the route that is currently mounted against
// the committed path, and applies their decision to it. It is what an
// application calls after the authority a guard reads changes under a page the
// user is already on: install the new snapshot first, then revalidate, and the
// guards stay the single declarative place where access is decided.
//
// A route its guards still allow is left untouched. One they now refuse is torn
// down like the navigation that would have been refused: the routed component
// is unmounted and dropped. A guard that redirects navigates to its
// destination. With no routed component mounted, or none for the committed
// path, nothing happens, so calling it again after a refusal is a no-op.
//
// It must not be called from a guard: a guard returns a decision, and the
// router is what acts on it.
func Revalidate() {
	_ = RevalidateContext(context.Background())
}

// RevalidateContext revalidates the mounted route like Revalidate and returns
// the outcome: nil when the route may stay, when a redirect committed, or when
// there was nothing to revalidate, and otherwise the error the guards decided
// on (ErrNavigationForbidden, ErrNavigationBlocked, ErrInvalidGuardResult), the
// *HostNavigationError of a host handoff, or the context error.
func RevalidateContext(parent context.Context) error {
	// A caller with no context to give passes none, the way Revalidate does.
	// Normalized here, at the one entry point: a redirect derives its context
	// from this parent, and deriving from a nil one panics.
	if parent == nil {
		parent = context.Background()
	}
	// Nothing mounted is nothing to revalidate: a navigation that has not
	// committed yet still has its own guards ahead of it, and a refusal already
	// cleared the owner it dropped.
	if currentComponent == nil {
		return nil
	}
	r, guards, params := matchCurrentRoute(activePathSig.Get())
	if r == nil {
		return nil
	}
	// The match only speaks for the route that actually committed what is
	// mounted. It does not when a route is registered for the path the
	// not-found component is sitting on: those guards never ran to put anything
	// there, so they have no say over a component that is not theirs.
	if r.component != currentComponent {
		return nil
	}
	// Taking the navigation generation cancels whatever is in flight: it was
	// authorized under the snapshot this call replaces, so its loader must not
	// come back and commit over the decision made here.
	ctx, _ := beginNavigation(parent)
	if err := ctx.Err(); err != nil {
		return err
	}

	guardResult, guardErr := runGuards(guards, params)
	if guardErr != nil {
		// Every refusal fails closed the same way, a legacy guard that blocks
		// included: the route is mounted, so leaving it there is the one outcome
		// a refusal cannot have. Navigation to a refused destination keeps its
		// own behavior, where nothing was committed to take back.
		revokeMountedRoute(guardErr)
		return guardErr
	}
	switch guardResult.Action {
	case GuardAllow:
		settleRevalidatedRoute()
		return nil
	case GuardHostReplace:
		// There is no document to replace here and the host page is not a route
		// this router could load in its place, so the handoff is reported with
		// its target and the mounted route is left alone, exactly as
		// NavigateContext reports it.
		hostErr := &HostNavigationError{Path: guardResult.Path}
		failNavigation(hostErr)
		return hostErr
	}

	redirectCtx, err := nextRedirectContext(parent)
	if err != nil {
		failNavigation(err)
		return err
	}
	// Taken before the redirect, released after it commits: the destination is
	// what decides whether this route is still the one the user owns.
	owner := currentComponent
	// The stub keeps no browser history, so a push and a replace redirect both
	// navigate, as they do for a refused destination.
	if err := NavigateContext(redirectCtx, guardResult.Path); err != nil {
		// The destination refused, failed to load or was cancelled: nothing
		// committed over the route the user is on, so it stays current and keeps
		// everything it holds.
		return err
	}
	releaseReplacedOwner(owner)
	return nil
}

// releaseReplacedOwner runs the lifecycle of the routed owner a committed
// redirect left behind. Navigation outside the browser has no DOM to tear down
// and only moves the pointer to the routed owner, so the one it replaced is
// released here, exactly once, where the browser gets the same teardown from
// navigateImpl. An owner the destination resolved back to is still current and
// is left running.
func releaseReplacedOwner(owner core.Component) {
	if owner == nil || owner == currentComponent {
		return
	}
	tearDownRoutedOwner(owner)
}

// revokeMountedRoute drops the routed owner its guards just refused, and what
// it loaded with it, so nothing reads its data or its metadata after the
// refusal. The committed path is kept, so a repeated revalidation finds nothing
// mounted and does nothing, while an explicit navigation can enter the route
// again once its guards allow it.
func revokeMountedRoute(err error) {
	owner := currentComponent
	currentComponent = nil
	tearDownRoutedOwner(owner)
	revokeCommittedRoute(err)
}

// tearDownRoutedOwner runs the lifecycle of a routed owner that is no longer
// current, exactly once.
func tearDownRoutedOwner(owner core.Component) {
	core.TriggerUnmount(owner)
	if unmounter, ok := owner.(routeUnmounter); ok {
		unmounter.Unmount()
	}
}
