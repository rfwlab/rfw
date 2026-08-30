//go:build js && wasm

package router

import (
	"context"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	js "github.com/rfwlab/rfw/v2/js"
)

// Revalidate re-runs the guards of the route that is currently mounted against
// the committed path, and applies their decision to it. It is what an
// application calls after the authority a guard reads changes under a page the
// user is already on: install the new snapshot first, then revalidate, and the
// guards stay the single declarative place where access is decided.
//
// A route its guards still allow is left untouched. One they now refuse is torn
// down like the navigation that would have been refused: the routed component
// unmounts, its subtree leaves the DOM, and the shell around the outlet stays
// mounted. A guard that redirects sends navigation to its destination, and one
// that hands off to the host loads the host page. With no routed component
// mounted, or none for the committed path, nothing happens, so calling it again
// after a refusal is a no-op.
//
// It must not be called from a guard: a guard returns a decision, and the
// router is what acts on it.
func Revalidate() {
	_ = RevalidateContext(context.Background())
}

// RevalidateContext revalidates the mounted route like Revalidate and returns
// the outcome: nil when the route may stay, when a redirect committed, when the
// browser was handed to the host, or when there was nothing to revalidate, and
// otherwise the error the guards decided on (ErrNavigationForbidden,
// ErrNavigationBlocked, ErrInvalidGuardResult) or the context error.
func RevalidateContext(parent context.Context) error {
	// A caller with no context to give passes none, the way Revalidate does.
	// Normalized here, at the one entry point: a redirect derives its context
	// from this parent, and deriving from a nil one panics.
	if parent == nil {
		parent = context.Background()
	}
	var revalidationErr error
	core.TryNavigate(activePathSig.Get(), func() {
		revalidationErr = revalidateImpl(parent)
	})
	return revalidationErr
}

func revalidateImpl(parent context.Context) error {
	// Nothing mounted is nothing to revalidate: a navigation that has not
	// committed yet still has its own guards ahead of it, and a refusal already
	// cleared the owner it tore down.
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
		// The host owns the destination and the browser loads it as a document.
		// Nothing commits and nothing is torn down first: the page the user is
		// on stays exactly as it is until the browser unloads it.
		hostReplace(guardResult.Path)
		return nil
	}

	redirectCtx, err := nextRedirectContext(parent)
	if err != nil {
		failNavigation(err)
		return err
	}
	// The entry the browser is sitting on is the one the guard just refused, so
	// both redirect actions replace it: pushing would leave the refused path
	// behind for back to return to, where the guard refuses it again. From
	// there it is ordinary routing, which unmounts the current owner once.
	return navigateImpl(redirectCtx, guardResult.Path, historyReplace)
}

// revokeMountedRoute tears down the routed owner its guards just refused. Only
// the route goes: the persistent root mounted with MountRoot and the outlet
// inside it stay where they are, so the next allowed navigation renders into
// the shell the user is still looking at. What the route loaded goes with it,
// so nothing reads its data or its metadata after the refusal. The committed
// path is kept, so a repeated revalidation finds nothing mounted and does
// nothing, while an explicit navigation can enter the route again once its
// guards allow it.
func revokeMountedRoute(err error) {
	owner := currentComponent
	// Dropped before the teardown, not after: the outlet repaints whatever
	// currentComponent points at whenever a component around it paints, and
	// unmounting paints.
	currentComponent = nil
	core.Log().Debug("Unmounting revoked component: %s", owner.GetName())
	core.TriggerUnmount(owner)
	// Unmount first, so the component's own cleanup, its scope and its host
	// registrations still see the tree they own.
	owner.Unmount()
	clearRoutedSubtree(owner.GetID())
	revokeCommittedRoute(err)
}

// clearRoutedSubtree removes the DOM the routed owner had, and nothing else.
// With a live outlet the marker div is the router's anchor inside the shell, so
// its content goes and the marker itself stays where a later navigation and a
// shell re-render expect to find it. Without one the router rendered the page
// into #app itself and the owner's own root is removed.
//
// dom.ComponentRoot answers #app for an id it cannot resolve, which here would
// mean erasing the shell instead of the route, so a resolved root is used only
// when it really carries the id it was asked for.
func clearRoutedSubtree(id string) {
	if content, ok := routedOutletContent(); ok {
		content.SetHTML("")
		return
	}
	root := dom.ComponentRoot(id)
	if root.IsNull() || root.IsUndefined() || root.Attr("data-component-id") != id {
		return
	}
	root.Call("remove")
}

// routedOutletContent resolves the element a live outlet renders routed
// components into: the marker div, or the outlet's own root when the template
// carries no marker. It reports false when there is no outlet or its root is
// not in the document, the case renderChild falls back to #app for.
func routedOutletContent() (dom.Element, bool) {
	if liveOutlet == nil {
		return dom.Element{Value: js.Null()}, false
	}
	root := dom.ComponentRoot(liveOutlet.GetID())
	if root.IsNull() || root.IsUndefined() || root.Attr("data-component-id") != liveOutlet.GetID() {
		return dom.Element{Value: js.Null()}, false
	}
	if target := root.Query("[data-router-outlet]"); !target.IsNull() && !target.IsUndefined() {
		return target, true
	}
	return root, true
}
