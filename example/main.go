//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/example/components"
	"github.com/rfwlab/rfw/v1/router"
	"github.com/rfwlab/rfw/v1/state"
)

func main() {
	store := state.NewStore("default")
	store.Set("count", 0)
	router.ExposeNavigate()
	state.ExposeUpdateStore()

	mainComponent := components.NewMainComponent()
	testComponent := components.NewTestComponent()
	anotherComponent := components.NewAnotherComponent()
	eventComponent := components.NewEventComponent()

	router.RegisterRoute("/", mainComponent)
	router.RegisterRoute("/test", testComponent)
	router.RegisterRoute("/user/:name", anotherComponent)
	router.RegisterRoute("/event", eventComponent)

	router.InitRouter()

	select {}
}
