//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/example/components"
	"github.com/rfwlab/rfw/example/plugins/logging"
	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/router"
	"github.com/rfwlab/rfw/v1/state"
)

func main() {
	core.RegisterPlugin(logging.New())
	store := state.NewStore("default", state.WithModule("app"))
	store.Set("count", 0)
	store.Set("allowProtected", false)
	if store.Get("sharedState") == nil {
		store.Set("sharedState", "Initial State")
	}

	testingStore := state.NewStore("testing")
	if testingStore.Get("testingState") == nil {
		testingStore.Set("testingState", "Testing Initial State")
	}

	router.ExposeNavigate()
	state.ExposeUpdateStore()

	authGuard := func(params map[string]string) bool {
		allowed, _ := store.Get("allowProtected").(bool)
		return allowed
	}

	router.RegisterRoute(router.Route{
		Path:      "/",
		Component: func() core.Component { return components.NewMainComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/test",
		Component: func() core.Component { return components.NewTestComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/user/:name",
		Component: func() core.Component { return components.NewAnotherComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/event",
		Component: func() core.Component { return components.NewEventComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/computed",
		Component: func() core.Component { return components.NewComputedComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/watcher",
		Component: func() core.Component { return components.NewWatcherComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/stores",
		Component: func() core.Component { return components.NewStoresComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/parent",
		Component: func() core.Component { return components.NewParentComponent() },
		Children: []router.Route{
			{
				Path:      "/parent/child",
				Component: func() core.Component { return components.NewChildComponent() },
			},
		},
	})
	router.RegisterRoute(router.Route{
		Path:      "/protected",
		Component: func() core.Component { return components.NewProtectedComponent() },
		Guards:    []router.Guard{authGuard},
	})

	router.InitRouter()

	select {}
}
