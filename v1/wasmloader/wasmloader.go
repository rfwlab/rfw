//go:build js && wasm

package wasmloader

import (
	"fmt"
	"strings"

	"github.com/rfwlab/rfw/v1/js"
)

type Options struct {
	Go         js.Value
	Color      string
	Height     string
	Blur       string
	SkipLoader bool
}

func candidateURLs(url string) []string {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return nil
	}

	base := trimmed
	query := ""
	if idx := strings.Index(trimmed, "?"); idx != -1 {
		base = trimmed[:idx]
		query = trimmed[idx:]
	}

	var urls []string
	if strings.HasSuffix(base, ".wasm") && !strings.HasSuffix(base, ".wasm.br") {
		urls = append(urls, base+".br"+query)
	}
	urls = append(urls, trimmed)
	return urls
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

func finishBar(bar, intervalID js.Value, intervalFn js.Func) {
	if !bar.Truthy() {
		return
	}
	js.Call("clearInterval", intervalID)
	intervalFn.Release()
	bar.Get("style").Set("width", "100%")
	removeFn := js.Func{}
	removeFn = js.FuncOf(func(this js.Value, args []js.Value) any {
		bar.Call("remove")
		removeFn.Release()
		return nil
	})
	js.Call("setTimeout", removeFn, 300)
}

func removeBar(bar, intervalID js.Value, intervalFn js.Func) {
	if !bar.Truthy() {
		return
	}
	js.Call("clearInterval", intervalID)
	intervalFn.Release()
	bar.Call("remove")
}

func instantiate(resp js.Value, opts Options, bar, intervalID js.Value, intervalFn js.Func) {
	arrayBufFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		bytes := args[0]
		finishBar(bar, intervalID, intervalFn)
		wasm := js.WebAssembly()
		instantiateFn := js.Func{}
		instantiateFn = js.FuncOf(func(this js.Value, args []js.Value) any {
			inst := args[0].Get("instance")
			opts.Go.Call("run", inst)
			instantiateFn.Release()
			return nil
		})
		wasm.Call("instantiate", bytes, opts.Go.Get("importObject")).Call("then", instantiateFn)
		arrayBufFn.Release()
		return nil
	})
	resp.Call("arrayBuffer").Call("then", arrayBufFn)
}

func Load(url string, opts Options) {
	var bar js.Value
	var intervalID js.Value
	var intervalFn js.Func
	if !opts.SkipLoader {
		bar, intervalID, intervalFn = createBar(opts)
	}

	urls := candidateURLs(url)
	if len(urls) == 0 {
		return
	}

	var tryFetch func(int)
	tryFetch = func(idx int) {
		if idx >= len(urls) {
			removeBar(bar, intervalID, intervalFn)
			js.Console().Call("error", fmt.Sprintf("failed to load wasm bundle from candidates: %s", strings.Join(urls, ", ")))
			return
		}

		current := urls[idx]
		success := js.Func{}
		failure := js.Func{}

		success = js.FuncOf(func(this js.Value, args []js.Value) any {
			resp := args[0]
			if !resp.Get("ok").Bool() {
				success.Release()
				failure.Release()
				tryFetch(idx + 1)
				return nil
			}
			success.Release()
			failure.Release()
			instantiate(resp, opts, bar, intervalID, intervalFn)
			return nil
		})

		failure = js.FuncOf(func(this js.Value, args []js.Value) any {
			success.Release()
			failure.Release()
			tryFetch(idx + 1)
			return nil
		})

		js.Fetch(current).Call("then", success, failure)
	}

	tryFetch(0)
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
