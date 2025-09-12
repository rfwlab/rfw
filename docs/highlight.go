//go:build js && wasm

package main

import (
	js "github.com/rfwlab/rfw/v1/js"
	hljs "github.com/rfwlab/rfw/v1/js/shim/highlightjs"
)

func init() {
	hljs.RegisterLanguage("rtml", func(h js.Value) js.Value {
		xml := h.Call("getLanguage", "xml")
		reg := js.Get("RegExp")
		interpolation := js.ValueOf(map[string]any{
			"className": "template-variable",
			"begin":     reg.New("\\{"),
			"end":       reg.New("\\}"),
			"relevance": 0,
		})
		arr := js.NewArray()
		arr.Push(interpolation)
		contains := js.ArrayOf(xml.Get("contains")).Concat(arr).Value
		rtml := h.Call("inherit", xml, map[string]any{"contains": contains})
		return rtml
	})
}
