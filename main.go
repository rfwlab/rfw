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

	framework.RegisterRoute("/", mainComponent)
	framework.RegisterRoute("/test", myComponent)
	framework.RegisterRoute("/another", anotherComponent)

	framework.InitRouter()
	framework.Navigate("/")

	select {}
}
