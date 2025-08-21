//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"

	"github.com/rfwlab/rfw/v1/dom"
	js "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/state"
)

func main() {
	document := js.Document()
	dataEl := document.Call("getElementById", "__RFWDATA__")
	var props map[string]any
	if dataEl.Truthy() {
		if err := json.Unmarshal([]byte(dataEl.Get("textContent").String()), &props); err != nil {
			props = map[string]any{}
		}
	}

	store := state.NewStore("default", state.WithModule("app"))
	if v, ok := props["count"].(float64); ok {
		store.Set("count", int(v))
	} else {
		store.Set("count", 0)
	}

	root := document.Call("getElementById", "app")
	span := root.Call("querySelector", "[data-store='app.default.count']")
	store.OnChange("count", func(v any) {
		span.Set("textContent", fmt.Sprintf("%v", v))
	})

	dom.RegisterHandlerFunc("increment", func() {
		if val, ok := store.Get("count").(int); ok {
			store.Set("count", val+1)
		}
	})

	dom.BindStoreInputs(root)
	dom.BindEventListeners("", root)

	select {}
}
