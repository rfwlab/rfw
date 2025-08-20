//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/docs/components"
	"github.com/rfwlab/rfw/v1/core"
	docplug "github.com/rfwlab/rfw/v1/plugins/docs"
	"github.com/rfwlab/rfw/v1/router"
	"github.com/rfwlab/rfw/v1/state"
)

func main() {
	// Establish a default application store so components can initialize
	// without explicitly providing one.
	state.NewStore("default", state.WithModule("app"))

	core.RegisterPlugin(docplug.New("/articles/sidebar.json"))
	router.RegisterRoute(router.Route{
		Path:      "/",
		Component: func() core.Component { return components.NewHomeComponent() },
	})
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
	router.InitRouter()
	select {}
}
