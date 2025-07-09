//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/components"
	"github.com/rfwlab/rfw/framework"
)

func main() {
	framework.NewStore("default")

	mainComponent := components.NewMainComponent()
	testComponent := components.NewTestComponent()
	anotherComponent := components.NewAnotherComponent()

	framework.RegisterRoute("/", mainComponent)
	framework.RegisterRoute("/test", testComponent)
	framework.RegisterRoute("/user/:name", anotherComponent)

	framework.InitRouter()

	select {}
}
