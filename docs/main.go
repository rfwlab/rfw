//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/docs/components"
	excomponents "github.com/rfwlab/rfw/docs/examples/components"
	plugs "github.com/rfwlab/rfw/docs/examples/plugins"
	"github.com/rfwlab/rfw/docs/examples/plugins/logger"
	mon "github.com/rfwlab/rfw/docs/examples/plugins/monitor"
	soccer "github.com/rfwlab/rfw/docs/examples/plugins/soccer"
	_ "github.com/rfwlab/rfw/docs/pages"
	"github.com/rfwlab/rfw/v1/core"
	docplug "github.com/rfwlab/rfw/v1/plugins/docs"
	highlight "github.com/rfwlab/rfw/v1/plugins/highlight"
	"github.com/rfwlab/rfw/v1/plugins/i18n"
	"github.com/rfwlab/rfw/v1/router"
	"github.com/rfwlab/rfw/v1/state"
)

func main() {
	// Enable dev mode for the examples.
	core.SetDevMode(true)

	// Establish application stores used by the examples.
	store := state.NewStore("default", state.WithModule("app"), state.WithPersistence(), state.WithDevTools())
	store.Set("count", 0)
	store.Set("allowProtected", false)
	store.Set("apiData", "")
	if store.Get("sharedState") == nil {
		store.Set("sharedState", "Initial State")
	}

	testingStore := state.NewStore("testing")
	if testingStore.Get("testingState") == nil {
		testingStore.Set("testingState", "Testing Initial State")
	}

	router.ExposeNavigate()
	state.ExposeUpdateStore()

	core.RegisterPlugin(logger.New())
	core.RegisterPlugin(i18n.New(map[string]map[string]string{
		"en": {"hello": "Hello"},
		"it": {"hello": "Ciao"},
	}))
	core.RegisterPlugin(mon.New())
	core.RegisterPlugin(highlight.New())
	core.RegisterPlugin(soccer.New())
	core.RegisterPlugin(docplug.New("/articles/sidebar.json"))

	router.NotFoundComponent = func() core.Component { return components.NewNotFoundComponent() }

	router.RegisterRoute(router.Route{
		Path:      "/docs",
		Component: func() core.Component { return components.NewDocsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/docs/:page",
		Component: func() core.Component { return components.NewDocsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/docs/:section/:page",
		Component: func() core.Component { return components.NewDocsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/ssc",
		Component: func() core.Component { return components.NewSSCComponent() },
	})
	// Example routes mounted under /examples
	router.RegisterRoute(router.Route{
		Path:      "/examples/main",
		Component: func() core.Component { return excomponents.NewMainComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/test",
		Component: func() core.Component { return excomponents.NewTestComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/dynamic",
		Component: func() core.Component { return excomponents.NewDynamicComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/slots",
		Component: func() core.Component { return excomponents.NewSlotParentComponent(nil) },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/user/:name",
		Component: func() core.Component { return excomponents.NewAnotherComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/params/:id",
		Component: func() core.Component { return excomponents.NewParamsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/event",
		Component: func() core.Component { return excomponents.NewEventComponent() },
		Children: []router.Route{
			{
				Path:      "/examples/event/listener",
				Component: func() core.Component { return excomponents.NewEventListenerComponent() },
			},
			{
				Path:      "/examples/event/observer",
				Component: func() core.Component { return excomponents.NewObserverComponent() },
			},
		},
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/computed",
		Component: func() core.Component { return excomponents.NewComputedComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/webgl",
		Component: func() core.Component { return excomponents.NewWebGLComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/animations",
		Component: func() core.Component { return excomponents.NewAnimationComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/input",
		Component: func() core.Component { return excomponents.NewInputComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/cinema",
		Component: func() core.Component { return excomponents.NewCinemaComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/plugins",
		Component: func() core.Component { return plugs.NewPluginsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/plugin-directives",
		Component: func() core.Component { return excomponents.NewPluginDirectivesComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/stores",
		Component: func() core.Component { return excomponents.NewStoresComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/parent",
		Component: func() core.Component { return excomponents.NewParentComponent() },
		Children: []router.Route{
			{
				Path:      "/examples/parent/child",
				Component: func() core.Component { return excomponents.NewChildComponent() },
			},
		},
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/protected",
		Component: func() core.Component { return excomponents.NewProtectedComponent() },
		Guards: []router.Guard{
			func(params map[string]string) bool {
				allowed, _ := store.Get("allowProtected").(bool)
				return allowed
			},
		},
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/complex/:user/:section",
		Component: func() core.Component { return excomponents.NewComplexRoutingComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/state",
		Component: func() core.Component { return excomponents.NewStateManagementComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/state-bindings",
		Component: func() core.Component { return excomponents.NewStateBindingsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/api",
		Component: func() core.Component { return excomponents.NewAPIIntegrationComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/fetchjson",
		Component: func() core.Component { return excomponents.NewFetchJSONComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/twitch/login",
		Component: func() core.Component { return excomponents.NewTwitchLoginComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/twitch/callback",
		Component: func() core.Component { return excomponents.NewTwitchCallbackComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/signal-bindings",
		Component: func() core.Component { return excomponents.NewSignalBindingsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/signals",
		Component: func() core.Component { return excomponents.NewSignalsEffectsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/pathfinding",
		Component: func() core.Component { return excomponents.NewPathfindingComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/examples/netcode",
		Component: func() core.Component { return excomponents.NewNetcodeComponent() },
	})

	router.InitRouter()
	select {}
}
