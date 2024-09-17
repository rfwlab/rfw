// Path: /home/mirko/Projects/personal/rfw/main.go
//go:build js && wasm

package main

import (
	"github.com/mirkobrombin/rfw/components"
	"github.com/mirkobrombin/rfw/framework"
)

func main() {
	framework.NewStore("default")

	mainComponent := components.NewMainComponent()
	testComponent := components.NewTestComponent()
	anotherComponent := components.NewAnotherComponent()

	framework.RegisterRoute("/", mainComponent)
	framework.RegisterRoute("/test", testComponent)
	framework.RegisterRoute("/another", anotherComponent)

	framework.InitRouter()

	select {}
}
