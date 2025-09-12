//go:build js && wasm

package wasmloader

import (
	"fmt"

	"github.com/rfwlab/rfw/v1/js"
)

type Options struct {
	Go         js.Value
	Color      string
	Height     string
	Blur       string
	SkipLoader bool
}

func createBar(opts Options) (js.Value, js.Value, js.Func) {
	doc := js.Document()
	bar := doc.Call("createElement", "div")
	color := opts.Color
	if color == "" {
		color = "#ff0000"
	}
	blur := opts.Blur
	if blur == "" {
		blur = "8px"
	}
	height := opts.Height
	if height == "" {
		height = "4px"
	}
	style := bar.Get("style")
	style.Set("position", "fixed")
	style.Set("top", "0")
	style.Set("left", "0")
	style.Set("width", "0%")
	style.Set("height", height)
	style.Set("background", color)
	style.Set("boxShadow", fmt.Sprintf("0 0 %s %s", blur, color))
	style.Set("zIndex", 9999)
	style.Set("transition", "width 0.3s ease")
	doc.Get("body").Call("appendChild", bar)

	progress := 0.0
	intervalFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		progress += js.Math().Call("random").Float() * 10
		if progress > 90 {
			progress = 90
		}
		bar.Get("style").Set("width", fmt.Sprintf("%f%%", progress))
		return nil
	})
	intervalID := js.Call("setInterval", intervalFn, 200)
	return bar, intervalID, intervalFn
}

func Load(url string, opts Options) {
	var bar js.Value
	var intervalID js.Value
	var intervalFn js.Func
	if !opts.SkipLoader {
		bar, intervalID, intervalFn = createBar(opts)
	}

	js.Fetch(url).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		resp := args[0]
		return resp.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			bytes := args[0]
			if bar.Truthy() {
				js.Call("clearInterval", intervalID)
				intervalFn.Release()
				bar.Get("style").Set("width", "100%")
				js.Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
					bar.Call("remove")
					return nil
				}), 300)
			}
			wasm := js.WebAssembly()
			wasm.Call("instantiate", bytes, opts.Go.Get("importObject")).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				inst := args[0].Get("instance")
				opts.Go.Call("run", inst)
				return nil
			}))
			return nil
		}))
	}))
}

func loadFunc(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return nil
	}
	url := args[0].String()
	var opt Options
	if len(args) > 1 {
		obj := args[1]
		opt.Go = obj.Get("go")
		opt.Color = obj.Get("color").String()
		opt.Height = obj.Get("height").String()
		opt.Blur = obj.Get("blur").String()
		opt.SkipLoader = obj.Get("skipLoader").Bool()
	}
	Load(url, opt)
	return nil
}

func init() {
	js.Set("WasmLoader", map[string]any{
		"load": js.FuncOf(loadFunc),
	})
}
