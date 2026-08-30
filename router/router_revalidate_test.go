package router

import (
	"context"
	"errors"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
)

// revalidatePage is a routed component that counts the lifecycle a revocation
// has to run exactly once, and owns a scope the way a page with running work
// does.
type revalidatePage struct {
	name      string
	scope     *core.Scope
	unmounted int
	paramRuns int
	cleanups  int
}

func newRevalidatePage(name string) *revalidatePage {
	page := &revalidatePage{name: name}
	page.scope = core.NewScope()
	page.scope.Defer(func() { page.cleanups++ })
	return page
}

func (p *revalidatePage) Render() string          { return "" }
func (p *revalidatePage) Mount()                  {}
func (p *revalidatePage) OnMount()                {}
func (p *revalidatePage) OnUnmount()              {}
func (p *revalidatePage) GetName() string         { return p.name }
func (p *revalidatePage) GetID() string           { return p.name }
func (p *revalidatePage) SetSlots(map[string]any) {}
func (p *revalidatePage) IsMounted() bool         { return p.unmounted == 0 }

func (p *revalidatePage) Unmount() {
	p.unmounted++
	p.scope.Close()
}

func (p *revalidatePage) OnParams(map[string]string) { p.paramRuns++ }

func mustRevalidatePage(t *testing.T, c core.Component) *revalidatePage {
	t.Helper()
	page, ok := c.(*revalidatePage)
	if !ok {
		t.Fatalf("expected *revalidatePage, got %T", c)
	}
	return page
}

// sessionRoute registers a route whose guard reads a decision the test changes
// between navigations, the shape of an application that installs a new
// authorization snapshot and then revalidates. It counts every guard run, every
// component creation and every loader run.
type sessionRoute struct {
	decide  func(map[string]string) GuardResult
	guards  int
	created int
	loaded  int
}

func registerSessionRoute(t *testing.T, path string, decide func(map[string]string) GuardResult) *sessionRoute {
	t.Helper()
	sr := &sessionRoute{decide: decide}
	RegisterRoute(Route{
		Path: path,
		ResultGuards: []ResultGuard{func(params map[string]string) GuardResult {
			sr.guards++
			return sr.decide(params)
		}},
		Component: func() core.Component {
			sr.created++
			return newRevalidatePage("session")
		},
		Loader: func(context.Context, LoadContext) (any, error) {
			sr.loaded++
			return "payload", nil
		},
		Meta: map[string]any{"section": "session"},
	})
	return sr
}

// A route its guards still allow is left exactly as it is: the guards run
// again, and nothing else does.
func TestRevalidateAllowedRouteIsANoOp(t *testing.T) {
	resetRouter(t)
	allowed := true
	session := registerSessionRoute(t, "/reval-open", func(map[string]string) GuardResult {
		if !allowed {
			return Forbid()
		}
		return Allow()
	})

	if err := NavigateContext(context.Background(), "/reval-open?tab=usage"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())
	guards, created, loaded, paramRuns := session.guards, session.created, session.loaded, page.paramRuns

	if err := RevalidateContext(context.Background()); err != nil {
		t.Fatalf("revalidate: %v", err)
	}

	if session.guards != guards+1 {
		t.Fatalf("guard ran %d times, want %d", session.guards, guards+1)
	}
	if session.created != created || session.loaded != loaded {
		t.Fatalf("revalidation re-ran the route: created=%d loaded=%d", session.created, session.loaded)
	}
	if page.paramRuns != paramRuns {
		t.Fatalf("revalidation re-ran OnParams %d times, want %d", page.paramRuns, paramRuns)
	}
	if page.unmounted != 0 || page.cleanups != 0 {
		t.Fatalf("revalidation tore the page down: unmounted=%d cleanups=%d", page.unmounted, page.cleanups)
	}
	if CurrentComponent() != core.Component(page) {
		t.Fatalf("current component changed to %#v", CurrentComponent())
	}
	if got := ActivePath().Get(); got != "/reval-open?tab=usage" {
		t.Fatalf("active path = %q", got)
	}
	if Status().Get() != NavigationReady || Error().Get() != nil {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
	if Data().Get() != "payload" {
		t.Fatalf("route data = %#v, want payload", Data().Get())
	}
	if Meta().Get()["section"] != "session" {
		t.Fatalf("route metadata = %#v", Meta().Get())
	}
}

// A guard that refuses a route the user is already on unmounts it once, and
// keeps the committed path so a later navigation can retry it.
func TestRevalidateForbidTearsDownTheOwnerOnce(t *testing.T) {
	resetRouter(t)
	allowed := true
	session := registerSessionRoute(t, "/reval-vault", func(map[string]string) GuardResult {
		if !allowed {
			return Forbid()
		}
		return Allow()
	})

	if err := NavigateContext(context.Background(), "/reval-vault"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())
	if Data().Get() != "payload" || Meta().Get()["section"] != "session" {
		t.Fatalf("the route committed no state to revoke: data=%#v meta=%#v", Data().Get(), Meta().Get())
	}

	allowed = false
	err := RevalidateContext(context.Background())

	if !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden revalidation, got %v", err)
	}
	// The refused route's loader data and metadata went with it: they described a
	// page the guards no longer allow.
	if Data().Get() != nil {
		t.Fatalf("the revoked route's data is still readable: %#v", Data().Get())
	}
	if meta := Meta().Get(); len(meta) != 0 {
		t.Fatalf("the revoked route's metadata is still readable: %#v", meta)
	}
	if page.unmounted != 1 {
		t.Fatalf("page unmounted %d times, want 1", page.unmounted)
	}
	if page.cleanups != 1 {
		t.Fatalf("scope cleanup ran %d times, want 1", page.cleanups)
	}
	if CurrentComponent() != nil {
		t.Fatalf("the revoked route stayed current: %#v", CurrentComponent())
	}
	if Status().Get() != NavigationError || !errors.Is(Error().Get(), ErrNavigationForbidden) {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
	if got := ActivePath().Get(); got != "/reval-vault" {
		t.Fatalf("active path = %q, want the committed path", got)
	}

	guards := session.guards
	if err := RevalidateContext(context.Background()); err != nil {
		t.Fatalf("repeated revalidation: %v", err)
	}
	if session.guards != guards {
		t.Fatalf("a repeated revalidation re-ran the guards: %d, want %d", session.guards, guards)
	}
	if page.unmounted != 1 || page.cleanups != 1 {
		t.Fatalf("a repeated revalidation tore the page down again: unmounted=%d cleanups=%d", page.unmounted, page.cleanups)
	}

	allowed = true
	if err := NavigateContext(context.Background(), "/reval-vault"); err != nil {
		t.Fatalf("navigating back into the route: %v", err)
	}
	if mustRevalidatePage(t, CurrentComponent()) == page {
		t.Fatal("the revoked page was remounted instead of a fresh one")
	}
}

// A legacy bool guard that now blocks tears the mounted route down the same
// way, while a blocked navigation to another route keeps its own behavior:
// nothing was committed there, so nothing is taken back.
func TestRevalidateLegacyBlockTearsDownAndNavigationIsUnchanged(t *testing.T) {
	resetRouter(t)
	allowed := true
	RegisterRoute(Route{
		Path:      "/reval-legacy",
		Guards:    []Guard{func(map[string]string) bool { return allowed }},
		Component: func() core.Component { return newRevalidatePage("legacy") },
		Loader:    func(context.Context, LoadContext) (any, error) { return "payload", nil },
		Meta:      map[string]any{"section": "legacy"},
	})
	blocked := 0
	RegisterRoute(Route{
		Path:      "/reval-legacy-other",
		Guards:    []Guard{func(map[string]string) bool { blocked++; return false }},
		Component: func() core.Component { return newRevalidatePage("other") },
	})
	RegisterRoute(Route{
		Path:         "/reval-legacy-forbidden",
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return Forbid() }},
		Component:    func() core.Component { return newRevalidatePage("forbidden") },
	})

	if err := NavigateContext(context.Background(), "/reval-legacy"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())

	// An ordinary denied navigation leaves the mounted page and the navigation
	// state exactly as they were.
	if err := NavigateContext(context.Background(), "/reval-legacy-other"); !errors.Is(err, ErrNavigationBlocked) {
		t.Fatalf("expected a blocked navigation, got %v", err)
	}
	if blocked != 1 {
		t.Fatalf("the blocking guard ran %d times, want 1", blocked)
	}
	if CurrentComponent() != core.Component(page) || page.unmounted != 0 {
		t.Fatalf("a blocked navigation disturbed the mounted page: current=%#v unmounted=%d", CurrentComponent(), page.unmounted)
	}
	if Status().Get() != NavigationReady || Error().Get() != nil {
		t.Fatalf("a blocked navigation changed navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
	if Data().Get() != "payload" || Meta().Get()["section"] != "legacy" {
		t.Fatalf("a blocked navigation dropped the mounted route's state: data=%#v meta=%#v", Data().Get(), Meta().Get())
	}
	// A refused destination is denied the same way: nothing committed there, so
	// the page the user is on keeps everything it loaded.
	if err := NavigateContext(context.Background(), "/reval-legacy-forbidden"); !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden navigation, got %v", err)
	}
	if CurrentComponent() != core.Component(page) || page.unmounted != 0 {
		t.Fatalf("a forbidden navigation disturbed the mounted page: current=%#v unmounted=%d", CurrentComponent(), page.unmounted)
	}
	if Data().Get() != "payload" || Meta().Get()["section"] != "legacy" {
		t.Fatalf("a forbidden navigation dropped the mounted route's state: data=%#v meta=%#v", Data().Get(), Meta().Get())
	}

	allowed = false
	err := RevalidateContext(context.Background())

	if !errors.Is(err, ErrNavigationBlocked) {
		t.Fatalf("expected a blocked revalidation, got %v", err)
	}
	if page.unmounted != 1 || page.cleanups != 1 {
		t.Fatalf("teardown ran wrong: unmounted=%d cleanups=%d", page.unmounted, page.cleanups)
	}
	if CurrentComponent() != nil {
		t.Fatalf("the blocked route stayed current: %#v", CurrentComponent())
	}
	if Status().Get() != NavigationError || !errors.Is(Error().Get(), ErrNavigationBlocked) {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
	if Data().Get() != nil {
		t.Fatalf("the revoked route's data is still readable: %#v", Data().Get())
	}
	if meta := Meta().Get(); len(meta) != 0 {
		t.Fatalf("the revoked route's metadata is still readable: %#v", meta)
	}
	if got := ActivePath().Get(); got != "/reval-legacy" {
		t.Fatalf("active path = %q, want the committed path", got)
	}
}

// A guard that redirects during revalidation routes to its destination the way
// navigation does, for both redirect actions.
func TestRevalidateRedirectRoutesToTheDestination(t *testing.T) {
	for name, decision := range map[string]GuardResult{
		"redirect": RedirectTo("/reval-login"),
		"replace":  ReplaceWith("/reval-login"),
	} {
		t.Run(name, func(t *testing.T) {
			resetRouter(t)
			result := decision
			allowed := true
			RegisterRoute(Route{Path: "/reval-login", Component: func() core.Component { return newRevalidatePage("login") }})
			session := registerSessionRoute(t, "/reval-account", func(map[string]string) GuardResult {
				if !allowed {
					return result
				}
				return Allow()
			})

			if err := NavigateContext(context.Background(), "/reval-account"); err != nil {
				t.Fatalf("navigate: %v", err)
			}
			page := mustRevalidatePage(t, CurrentComponent())
			created := session.created

			allowed = false
			if err := RevalidateContext(context.Background()); err != nil {
				t.Fatalf("revalidate: %v", err)
			}

			if got := mustRevalidatePage(t, CurrentComponent()).name; got != "login" {
				t.Fatalf("current component = %q, want login", got)
			}
			// The redirect committed over the refused route, so the owner it
			// replaced is released with it, once.
			if page.unmounted != 1 || page.cleanups != 1 {
				t.Fatalf("teardown ran %d times (cleanups %d), want 1", page.unmounted, page.cleanups)
			}
			if got := ActivePath().Get(); got != "/reval-login" {
				t.Fatalf("active path = %q, want /reval-login", got)
			}
			if session.created != created {
				t.Fatalf("the refused route was created again: %d", session.created)
			}
			if Status().Get() != NavigationReady {
				t.Fatalf("navigation status = %s, want ready", Status().Get())
			}
		})
	}
}

// A redirect that never commits leaves the route the user is on exactly as it
// is: the destination is what failed, so the owner still holds the page, its
// scope and everything hanging off it, and a later revalidation can decide
// again. Nothing may be released on the way to a destination that was refused.
func TestRevalidateFailedRedirectKeepsTheOwner(t *testing.T) {
	for name, register := range map[string]func(){
		"forbidden destination": func() {
			RegisterRoute(Route{
				Path:         "/reval-dead-login",
				Component:    func() core.Component { return newRevalidatePage("login") },
				ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return Forbid() }},
			})
		},
		"failing loader": func() {
			RegisterRoute(Route{
				Path:      "/reval-dead-login",
				Component: func() core.Component { return newRevalidatePage("login") },
				Loader: func(context.Context, LoadContext) (any, error) {
					return nil, errors.New("login unavailable")
				},
			})
		},
	} {
		t.Run(name, func(t *testing.T) {
			resetRouter(t)
			register()
			allowed := true
			registerSessionRoute(t, "/reval-dead", func(map[string]string) GuardResult {
				if !allowed {
					return ReplaceWith("/reval-dead-login")
				}
				return Allow()
			})

			if err := NavigateContext(context.Background(), "/reval-dead"); err != nil {
				t.Fatalf("navigate: %v", err)
			}
			page := mustRevalidatePage(t, CurrentComponent())

			allowed = false
			if err := RevalidateContext(context.Background()); err == nil {
				t.Fatal("a redirect to a destination that never commits reported success")
			}

			if CurrentComponent() != core.Component(page) {
				t.Fatalf("the owner was replaced by a redirect that failed: %#v", CurrentComponent())
			}
			if page.unmounted != 0 || page.cleanups != 0 {
				t.Fatalf("a failed redirect tore the owner down: unmounted=%d cleanups=%d", page.unmounted, page.cleanups)
			}
			if got := ActivePath().Get(); got != "/reval-dead" {
				t.Fatalf("active path = %q, want the committed path", got)
			}
		})
	}
}

// noContext returns the nil context a caller with none of its own passes. It
// goes through a function so the "do not pass a nil Context" analysis does not
// rewrite the very call under test.
func noContext() context.Context { return nil }

// A caller that passes no context gets the one Revalidate builds for itself,
// redirects included: a redirect derives its context from the parent, and
// deriving from a nil one is what the boundary normalizes away.
func TestRevalidateNilContextRedirects(t *testing.T) {
	resetRouter(t)
	allowed := true
	RegisterRoute(Route{Path: "/reval-nil-login", Component: func() core.Component { return newRevalidatePage("login") }})
	registerSessionRoute(t, "/reval-nil", func(map[string]string) GuardResult {
		if !allowed {
			return ReplaceWith("/reval-nil-login")
		}
		return Allow()
	})

	if err := NavigateContext(context.Background(), "/reval-nil"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())

	allowed = false
	if err := RevalidateContext(noContext()); err != nil {
		t.Fatalf("revalidate: %v", err)
	}

	if got := mustRevalidatePage(t, CurrentComponent()).name; got != "login" {
		t.Fatalf("current component = %q, want login", got)
	}
	if page.unmounted != 1 {
		t.Fatalf("the refused page unmounted %d times, want 1", page.unmounted)
	}
	if got := ActivePath().Get(); got != "/reval-nil-login" {
		t.Fatalf("active path = %q, want /reval-nil-login", got)
	}
}

// Two guards redirecting at each other during revalidation hit the same
// redirect-loop limit navigation does.
func TestRevalidateRedirectLoopFails(t *testing.T) {
	resetRouter(t)
	revalidating := false
	RegisterRoute(Route{
		Path:      "/reval-ping",
		Component: func() core.Component { return newRevalidatePage("ping") },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			if revalidating {
				return RedirectTo("/reval-pong")
			}
			return Allow()
		}},
	})
	RegisterRoute(Route{
		Path:         "/reval-pong",
		Component:    func() core.Component { return newRevalidatePage("pong") },
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult { return ReplaceWith("/reval-ping") }},
	})

	if err := NavigateContext(context.Background(), "/reval-ping"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	revalidating = true
	if err := RevalidateContext(context.Background()); !errors.Is(err, ErrRedirectLoop) {
		t.Fatalf("expected a redirect loop error, got %v", err)
	}
}

// With no routed component mounted there is nothing to revalidate.
func TestRevalidateWithoutAMountedRouteDoesNothing(t *testing.T) {
	resetRouter(t)
	session := registerSessionRoute(t, "/reval-idle", func(map[string]string) GuardResult { return Forbid() })

	if err := RevalidateContext(context.Background()); err != nil {
		t.Fatalf("revalidate: %v", err)
	}
	if session.guards != 0 {
		t.Fatalf("guards ran with nothing mounted: %d", session.guards)
	}
	if Status().Get() != NavigationIdle || Error().Get() != nil {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
}

// A committed path that matches no route (the not-found component) has no
// guards to re-run, so revalidation leaves it alone.
func TestRevalidateWithoutAMatchingRouteIsANoOp(t *testing.T) {
	resetRouter(t)
	session := registerSessionRoute(t, "/reval-present", func(map[string]string) GuardResult { return Forbid() })
	NotFoundComponent = func() core.Component { return newRevalidatePage("not-found") }

	if err := NavigateContext(context.Background(), "/reval-absent"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())

	if err := RevalidateContext(context.Background()); err != nil {
		t.Fatalf("revalidate: %v", err)
	}
	if session.guards != 0 {
		t.Fatalf("guards of another route ran: %d", session.guards)
	}
	if CurrentComponent() != core.Component(page) || page.unmounted != 0 {
		t.Fatalf("the not-found page was disturbed: current=%#v unmounted=%d", CurrentComponent(), page.unmounted)
	}
}

// A route registered for the path the not-found component is sitting on matches
// it without owning what is mounted: its guards never ran to commit anything
// there, so revalidation leaves the component alone instead of letting them
// decide the fate of a page that is not theirs.
func TestRevalidateIgnoresARouteThatDoesNotOwnTheMountedComponent(t *testing.T) {
	resetRouter(t)
	NotFoundComponent = func() core.Component { return newRevalidatePage("not-found") }

	if err := NavigateContext(context.Background(), "/reval-late"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())

	// The route arrives afterwards, the shape of a feature that registers its
	// routes when it is installed rather than at startup.
	session := registerSessionRoute(t, "/reval-late", func(map[string]string) GuardResult { return Forbid() })

	if err := RevalidateContext(context.Background()); err != nil {
		t.Fatalf("revalidate: %v", err)
	}
	if session.guards != 0 {
		t.Fatalf("the guards of a route that owns nothing ran %d times", session.guards)
	}
	if CurrentComponent() != core.Component(page) || page.unmounted != 0 {
		t.Fatalf("the mounted component was disturbed: current=%#v unmounted=%d", CurrentComponent(), page.unmounted)
	}
	if got := ActivePath().Get(); got != "/reval-late" {
		t.Fatalf("active path = %q, want the committed path", got)
	}
}

// A cancelled context is answered before the guards run, and the mounted route
// stays where it is.
func TestRevalidateCancelledContextSkipsGuards(t *testing.T) {
	resetRouter(t)
	allowed := true
	session := registerSessionRoute(t, "/reval-cancelled", func(map[string]string) GuardResult {
		if !allowed {
			return Forbid()
		}
		return Allow()
	})

	if err := NavigateContext(context.Background(), "/reval-cancelled"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())
	guards := session.guards

	allowed = false
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := RevalidateContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected a cancelled revalidation, got %v", err)
	}
	if session.guards != guards {
		t.Fatalf("guards ran for a cancelled revalidation: %d", session.guards)
	}
	if CurrentComponent() != core.Component(page) || page.unmounted != 0 {
		t.Fatalf("a cancelled revalidation tore the page down: current=%#v unmounted=%d", CurrentComponent(), page.unmounted)
	}
}

// A loader still running was authorized under the snapshot revalidation
// replaces: it must not commit afterwards, whether the mounted route survives
// the revalidation or not.
func TestRevalidateCancelsAnInFlightLoader(t *testing.T) {
	resetRouter(t)
	allowed := true
	registerSessionRoute(t, "/reval-current", func(map[string]string) GuardResult {
		if !allowed {
			return Forbid()
		}
		return Allow()
	})
	entered, release := make(chan struct{}, 2), make(chan struct{})
	pending := 0
	RegisterRoute(Route{
		Path: "/reval-pending",
		Component: func() core.Component {
			pending++
			return newRevalidatePage("pending")
		},
		Loader: func(_ context.Context, _ LoadContext) (any, error) {
			entered <- struct{}{}
			<-release
			return "late", nil
		},
	})

	if err := NavigateContext(context.Background(), "/reval-current"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())

	// The route is still allowed: the in-flight navigation is cancelled and the
	// mounted page stays.
	navigation := make(chan error, 1)
	go func() { navigation <- NavigateContext(context.Background(), "/reval-pending") }()
	<-entered

	if err := RevalidateContext(context.Background()); err != nil {
		t.Fatalf("revalidate: %v", err)
	}
	release <- struct{}{}
	if err := <-navigation; err == nil {
		t.Fatal("the cancelled navigation reported success")
	}
	if CurrentComponent() != core.Component(page) {
		t.Fatalf("a cancelled loader committed over the mounted route: %#v", CurrentComponent())
	}
	if pending != 0 {
		t.Fatalf("the cancelled destination was created %d times", pending)
	}
	if Status().Get() != NavigationReady || Error().Get() != nil {
		t.Fatalf("navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
	if Data().Get() != "payload" {
		t.Fatalf("route data = %#v, want the mounted route's", Data().Get())
	}

	// Now the guard refuses: the same cancelled loader must not resurrect the
	// route the revalidation tore down.
	go func() { navigation <- NavigateContext(context.Background(), "/reval-pending") }()
	<-entered
	allowed = false

	if err := RevalidateContext(context.Background()); !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden revalidation, got %v", err)
	}
	close(release)
	if err := <-navigation; err == nil {
		t.Fatal("the cancelled navigation reported success")
	}
	if CurrentComponent() != nil {
		t.Fatalf("a cancelled loader committed after the revocation: %#v", CurrentComponent())
	}
	if page.unmounted != 1 || page.cleanups != 1 {
		t.Fatalf("teardown ran %d times (cleanups %d), want 1", page.unmounted, page.cleanups)
	}
	if pending != 0 {
		t.Fatalf("the cancelled destination was created %d times", pending)
	}
}

// Revalidation is repeatable: an allowed route survives any number of them, and
// a refused one is torn down exactly once however often it is asked again.
func TestRevalidateIsIdempotentUnderRepetition(t *testing.T) {
	resetRouter(t)
	allowed := true
	session := registerSessionRoute(t, "/reval-stress", func(map[string]string) GuardResult {
		if !allowed {
			return Forbid()
		}
		return Allow()
	})

	if err := NavigateContext(context.Background(), "/reval-stress"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	page := mustRevalidatePage(t, CurrentComponent())

	const rounds = 50
	for i := 0; i < rounds; i++ {
		if err := RevalidateContext(context.Background()); err != nil {
			t.Fatalf("revalidation %d: %v", i, err)
		}
	}
	if session.guards != rounds+1 {
		t.Fatalf("guard ran %d times, want %d", session.guards, rounds+1)
	}
	if session.created != 1 || session.loaded != 1 || page.unmounted != 0 {
		t.Fatalf("repeated revalidation re-ran the route: created=%d loaded=%d unmounted=%d", session.created, session.loaded, page.unmounted)
	}

	allowed = false
	refusals := 0
	for i := 0; i < rounds; i++ {
		if err := RevalidateContext(context.Background()); err != nil {
			refusals++
		}
	}
	if refusals != 1 {
		t.Fatalf("%d revalidations refused, want 1", refusals)
	}
	if page.unmounted != 1 || page.cleanups != 1 {
		t.Fatalf("teardown ran %d times (cleanups %d), want 1", page.unmounted, page.cleanups)
	}
	if CurrentComponent() != nil {
		t.Fatalf("the revoked route stayed current: %#v", CurrentComponent())
	}
}
