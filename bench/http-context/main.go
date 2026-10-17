//go:build js && wasm

package main

import (
	"context"
	"encoding/hex"
	"time"

	rhttp "github.com/rfwlab/rfw/v2/http"
	"github.com/rfwlab/rfw/v2/js"
)

func main() {
	pending := map[string]context.CancelFunc{}
	js.Set("requestResults", js.Get("Array").New())
	js.ExposeFunc("requestProbe", func(_ js.Value, args []js.Value) any {
		mode, id, path, timeout := args[0].String(), args[1].String(), args[2].String(), args[3].Int()
		ctx, cancel := context.WithCancel(context.Background())
		if timeout > 0 {
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
		}
		pending[id] = cancel
		if mode == "precancelled" {
			cancel()
		}
		done := func(status int, body string) {
			delete(pending, id)
			cancel()
			item := js.Object().New()
			item.Set("id", id)
			item.Set("status", status)
			item.Set("body", body)
			js.Get("requestResults").Call("push", item)
		}
		if mode == "bytes" {
			rhttp.RequestBytesContext(ctx, path, rhttp.RequestOptions{}, func(status int, body []byte) { done(status, hex.EncodeToString(body)) })
		} else if mode == "legacy" {
			rhttp.Request(path, rhttp.RequestOptions{}, done)
		} else {
			rhttp.RequestContext(ctx, path, rhttp.RequestOptions{}, done)
		}
		return nil
	})
	js.ExposeFunc("cancelRequestProbe", func(_ js.Value, args []js.Value) any {
		if cancel := pending[args[0].String()]; cancel != nil {
			cancel()
		}
		return nil
	})
	js.Set("requestProbeReady", true)
	select {}
}
