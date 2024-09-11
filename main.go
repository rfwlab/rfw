//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/mirkobrombin/rfw/components"
	"github.com/mirkobrombin/rfw/framework"
)

var sharedState *framework.ReactiveVar

func ExposeStateUpdate() {
	framework.ExposeFunction("goUpdateState", func() {
		document := js.Global().Get("document")
		newValue := document.Call("getElementById", "stateInput").Get("value").String()
		sharedState.Set(newValue)
	})
}

func main() {
	sharedState = framework.NewReactiveVar("Initial State")
	sharedState.OnChange(func(newValue string) {
		framework.UpdateDOM(newValue)
	})

	myComponent := components.NewMyComponent()
	anotherComponent := components.NewAnotherComponent(sharedState)

	framework.ExposeNavigate()
	ExposeStateUpdate()

	framework.RegisterRoute("/", myComponent.Render)
	framework.RegisterRoute("/another", anotherComponent.Render)

	framework.InitRouter()

	framework.Navigate("/")

	select {}
}
