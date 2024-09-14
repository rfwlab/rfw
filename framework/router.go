//go:build js && wasm

package framework

import (
	"log"
	"syscall/js"
)

var routes = map[string]Component{}
var currentComponent Component

func RegisterRoute(path string, component Component) {
	routes[path] = component
}

func Navigate(path string) {
	if component, exists := routes[path]; exists {
		if currentComponent != nil {
			log.Println("Unmounting current component:", currentComponent.GetName())
			currentComponent.Unmount()
		}
		currentComponent = component
		UpdateDOM("", component.Render())
		js.Global().Get("history").Call("pushState", nil, "", path)
	}
}

func ExposeNavigate() {
	js.Global().Set("goNavigate", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		path := args[0].String()
		Navigate(path)
		return nil
	}))
}

func InitRouter() {
	js.Global().Get("window").Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		path := js.Global().Get("location").Get("pathname").String()
		Navigate(path)
		return nil
	}))
}
