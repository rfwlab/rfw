package router

import (
	"context"
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
)

type dataComponent struct {
	recordComponent
	data any
}

func (component *dataComponent) SetRouteData(data any) {
	component.data = data
}

func TestNamedRouteURLAndMetadata(t *testing.T) {
	resetRouter(t)
	RegisterRoute(Route{
		Path: "/teams/:team",
		Children: []Route{{
			Path:      "users/:user",
			Name:      "team-user",
			Component: func() core.Component { return &recordComponent{name: "user"} },
			Meta:      map[string]any{"title": "User"},
		}},
	})

	path, err := URL("team-user", map[string]string{
		"team": "core",
		"user": "Ada Lovelace",
	}, url.Values{"tab": {"activity"}})
	if err != nil {
		t.Fatalf("build URL: %v", err)
	}
	if path != "/teams/core/users/Ada%20Lovelace?tab=activity" {
		t.Fatalf("unexpected URL: %s", path)
	}
	if err := NavigateContext(context.Background(), "/teams/core/users/Ada%20Lovelace"); err != nil {
		t.Fatalf("navigate generated URL: %v", err)
	}
	component := CurrentComponent().(*recordComponent)
	if component.params["user"] != "Ada Lovelace" {
		t.Fatalf("route parameter was not decoded: %#v", component.params)
	}

	definitions := RegisteredRoutes()
	child := definitions[0].Children[0]
	if child.Name != "team-user" || child.Meta["title"] != "User" {
		t.Fatalf("named route metadata missing: %#v", child)
	}
}

func TestRouteLoaderCommitsDataAndMeta(t *testing.T) {
	resetRouter(t)
	var loadedContext LoadContext
	RegisterRoute(Route{
		Path: "/reports/:id",
		Name: "report",
		Component: func() core.Component {
			return &dataComponent{recordComponent: recordComponent{name: "report"}}
		},
		Loader: func(_ context.Context, loadContext LoadContext) (any, error) {
			loadedContext = loadContext
			return map[string]any{"total": 4}, nil
		},
		Meta: map[string]any{"section": "reports"},
	})

	if err := NavigateContext(context.Background(), "/reports/7?period=week"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	component := CurrentComponent().(*dataComponent)
	if !reflect.DeepEqual(component.data, map[string]any{"total": 4}) {
		t.Fatalf("loader data missing: %#v", component.data)
	}
	if loadedContext.Params["id"] != "7" || loadedContext.Query.Get("period") != "week" {
		t.Fatalf("loader context incorrect: %#v", loadedContext)
	}
	if Status().Get() != NavigationReady || Error().Get() != nil {
		t.Fatalf("unexpected navigation state: status=%s error=%v", Status().Get(), Error().Get())
	}
	if Meta().Get()["section"] != "reports" {
		t.Fatalf("route metadata missing: %#v", Meta().Get())
	}
}

func TestNewNavigationCancelsPreviousLoader(t *testing.T) {
	resetRouter(t)
	started := make(chan struct{})
	RegisterRoute(Route{
		Path:      "/slow",
		Component: func() core.Component { return &recordComponent{name: "slow"} },
		Loader: func(ctx context.Context, _ LoadContext) (any, error) {
			close(started)
			<-ctx.Done()
			return nil, ctx.Err()
		},
	})
	RegisterRoute(Route{
		Path:      "/fast",
		Component: func() core.Component { return &recordComponent{name: "fast"} },
	})

	result := make(chan error, 1)
	go func() {
		result <- NavigateContext(context.Background(), "/slow")
	}()
	<-started
	if err := NavigateContext(context.Background(), "/fast"); err != nil {
		t.Fatalf("fast navigation: %v", err)
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("slow loader was not cancelled: %v", err)
	}
	if CurrentComponent().GetName() != "fast" || Status().Get() != NavigationReady {
		t.Fatalf("stale loader replaced current route: component=%v status=%s", CurrentComponent(), Status().Get())
	}
}

func TestRouteRedirectInterpolatesParameters(t *testing.T) {
	resetRouter(t)
	RegisterRoute(Route{Path: "/legacy/:id", Redirect: "/users/:id"})
	RegisterRoute(Route{
		Path:      "/users/:id",
		Component: func() core.Component { return &dataComponent{recordComponent: recordComponent{name: "user"}} },
		Loader: func(ctx context.Context, load LoadContext) (any, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return load.Params["id"], nil
		},
	})

	if err := NavigateContext(context.Background(), "/legacy/42"); err != nil {
		t.Fatalf("redirect: %v", err)
	}
	component := CurrentComponent().(*dataComponent)
	if component.params["id"] != "42" || ActivePath().Get() != "/users/42" {
		t.Fatalf("redirect destination incorrect: component=%#v path=%s", component, ActivePath().Get())
	}
	if component.data != "42" {
		t.Fatalf("redirected loader did not complete: %#v", component.data)
	}
}

func TestRouteRedirectLoopFails(t *testing.T) {
	resetRouter(t)
	RegisterRoute(Route{Path: "/loop-a", Redirect: "/loop-b"})
	RegisterRoute(Route{Path: "/loop-b", Redirect: "/loop-a"})

	if err := NavigateContext(context.Background(), "/loop-a"); !errors.Is(err, ErrRedirectLoop) {
		t.Fatalf("expected redirect loop error, got %v", err)
	}
}

func TestCancelledNavigationDoesNotCommit(t *testing.T) {
	resetRouter(t)
	RegisterRoute(Route{
		Path:      "/cancelled",
		Component: func() core.Component { return &recordComponent{name: "cancelled"} },
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := NavigateContext(ctx, "/cancelled"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancelled context, got %v", err)
	}
	if CurrentComponent() != nil {
		t.Fatalf("cancelled navigation committed %#v", CurrentComponent())
	}
}
