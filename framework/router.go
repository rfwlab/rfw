//go:build js && wasm

package framework

import (
	"syscall/js"
)

var routes = map[string]func() string{}

func RegisterRoute(path string, renderFunc func() string) {
	routes[path] = renderFunc
}

func Navigate(path string) {
	if renderFunc, exists := routes[path]; exists {
		UpdateDOM(renderFunc())
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
