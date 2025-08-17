//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/example/components"
	plugs "github.com/rfwlab/rfw/example/plugins"
	"github.com/rfwlab/rfw/example/plugins/i18n"
	"github.com/rfwlab/rfw/example/plugins/logger"
	mon "github.com/rfwlab/rfw/example/plugins/monitor"
	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/router"
	"github.com/rfwlab/rfw/v1/state"
)

func main() {
	core.SetDevMode(true)
	store := state.NewStore("default", state.WithModule("app"), state.WithPersistence(), state.WithDevTools())
	store.Set("count", 0)
	store.Set("allowProtected", false)
	if store.Get("sharedState") == nil {
		store.Set("sharedState", "Initial State")
	}

	core.RegisterPlugin(logger.New())
	core.RegisterPlugin(i18n.New(map[string]map[string]string{
		"en": {
			"hello": "Hello",
		},
		"it": {
			"hello": "Ciao",
		},
	}))
	core.RegisterPlugin(mon.New())

	testingStore := state.NewStore("testing")
	if testingStore.Get("testingState") == nil {
		testingStore.Set("testingState", "Testing Initial State")
	}

	router.ExposeNavigate()
	state.ExposeUpdateStore()

	// Register a component that will be loaded on-demand using the `rt-is`
	// attribute inside templates. It renders a red cube with a large 'A'
	// at the center when referenced.
	core.RegisterComponent("red-cube", func() core.Component {
		return components.NewRedCubeComponent()
	})

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
		Path:      "/dynamic",
		Component: func() core.Component { return components.NewDynamicComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/slots",
		Component: func() core.Component { return components.NewSlotParentComponent(nil) },
	})
	router.RegisterRoute(router.Route{
		Path:      "/user/:name",
		Component: func() core.Component { return components.NewAnotherComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/event",
		Component: func() core.Component { return components.NewEventComponent() },
		Children: []router.Route{
			{
				Path:      "/event/listener",
				Component: func() core.Component { return components.NewEventListenerComponent() },
			},
			{
				Path:      "/event/observer",
				Component: func() core.Component { return components.NewObserverComponent() },
			},
		},
	})
	router.RegisterRoute(router.Route{
		Path:      "/computed",
		Component: func() core.Component { return components.NewComputedComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/plugins",
		Component: func() core.Component { return plugs.NewPluginsComponent() },
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
