//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/mirkobrombin/rfw/components"
	"github.com/mirkobrombin/rfw/framework"
)

func ExposeStateUpdate() {
	framework.ExposeFunction("goUpdateState", func(this js.Value, args []js.Value) interface{} {
		newValue := args[0].String()
		framework.GetStore("sharedStateStore").Set("sharedState", newValue)
		return nil
	})
}

func main() {
	framework.NewStore("sharedStateStore")

	mainComponent := components.NewMainComponent()
	myComponent := components.NewMyComponent()
	anotherComponent := components.NewAnotherComponent()

	framework.ExposeNavigate()
	ExposeStateUpdate()

	framework.RegisterRoute("/", mainComponent.Render)
	framework.RegisterRoute("/test", myComponent.Render)
	framework.RegisterRoute("/another", anotherComponent.Render)

	framework.InitRouter()
	framework.Navigate("/")

	select {}
}
