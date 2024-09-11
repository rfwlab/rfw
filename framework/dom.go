//go:build js && wasm

package framework

import (
	"syscall/js"
)

func UpdateDOM(html string) {
	document := js.Global().Get("document")
	app := document.Call("getElementById", "app")
	app.Set("innerHTML", html)
}
