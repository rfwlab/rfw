//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/docs/components"
	"github.com/rfwlab/rfw/docs/examples/plugins/logger"
	"github.com/rfwlab/rfw/docs/examples/plugins/monitor"
	"github.com/rfwlab/rfw/docs/examples/plugins/soccer"
	"github.com/rfwlab/rfw/docs/examples/registry"
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/plugins/docs"
	"github.com/rfwlab/rfw/v2/plugins/highlight"
	"github.com/rfwlab/rfw/v2/plugins/i18n"
	"github.com/rfwlab/rfw/v2/plugins/shortcut"
	"github.com/rfwlab/rfw/v2/plugins/toast"
	"github.com/rfwlab/rfw/v2/router"
	"github.com/rfwlab/rfw/v2/state"
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
	core.RegisterPlugin(monitor.New())
	core.RegisterPlugin(highlight.New())
	core.RegisterPlugin(toast.New())
	core.RegisterPlugin(soccer.New())
	core.RegisterPlugin(docs.New("/articles/sidebar.json"))
	core.RegisterPlugin(shortcut.New())

	router.NotFoundComponent = func() core.Component { return components.NewNotFoundComponent() }

	router.RegisterRoute(router.Route{
		Path:      "/",
		Component: func() core.Component { return components.NewDocsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/:page",
		Component: func() core.Component { return components.NewDocsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/:section/:page",
		Component: func() core.Component { return components.NewDocsComponent() },
	})
	router.RegisterRoute(router.Route{
		Path:      "/ssc",
		Component: func() core.Component { return components.NewSSCComponent() },
	})
	for _, route := range registry.Routes(store) {
		router.RegisterRoute(route)
	}

	router.InitRouter()
	select {}
}
