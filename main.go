// Path: /home/mirko/Projects/personal/rfw/main.go
//go:build js && wasm

package main

import (
	"github.com/mirkobrombin/rfw/components"
	"github.com/mirkobrombin/rfw/framework"
)

func main() {
	sharedStateStore := framework.NewStore("sharedStateStore")
	anotherStore := framework.NewStore("anotherStore")
	framework.GlobalStoreManager.RegisterStore("default", sharedStateStore)
	framework.GlobalStoreManager.RegisterStore("anotherStore", anotherStore)

	mainComponentA := components.NewMainComponent("A")
	mainComponentB := components.NewMainComponent("B")
	myComponent := components.NewMyComponent()
	anotherComponent := components.NewAnotherComponent()

	framework.ExposeNavigate()
	framework.ExposeUpdateStore()

	framework.RegisterRoute("/", mainComponentA)
	framework.RegisterRoute("/mainB", mainComponentB)
	framework.RegisterRoute("/test", myComponent)
	framework.RegisterRoute("/another", anotherComponent)

	framework.InitRouter()
	framework.Navigate("/")

	select {}
}
