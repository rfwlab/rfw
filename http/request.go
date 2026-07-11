//go:build js && wasm

package http

import "github.com/rfwlab/rfw/v2/js"

// RequestOptions configures a raw fetch performed by Request.
type RequestOptions struct {
	// Method is the HTTP method (defaults to GET when empty).
	Method string
	// Headers are extra request headers (e.g. Authorization, X-Workspace-ID).
	Headers map[string]string
	// Body is the request body for POST/PUT/PATCH (JSON string, etc.).
	Body string
}

// Request performs an uncached fetch with a custom method, headers and body and
// invokes cb with the HTTP status and the response body text once it resolves.
//
// Unlike FetchJSON/FetchText it does not cache and carries request headers, so
// it is the right primitive for authenticated and mutating requests (the caller
// supplies Authorization / workspace headers via RequestOptions.Headers). cb is
// invoked on the JS event loop; it may be nil.
func Request(url string, opts RequestOptions, cb func(status int, body string)) {
	o := js.Object().New()
	if opts.Method != "" {
		o.Set("method", opts.Method)
	}
	if len(opts.Headers) > 0 {
		h := js.Object().New()
		for k, v := range opts.Headers {
			h.Set(k, v)
		}
		o.Set("headers", h)
	}
	if opts.Body != "" {
		o.Set("body", opts.Body)
	}

	status := 0
	var onResp, onText, onErr js.Func
	onText = js.FuncOf(func(_ js.Value, a []js.Value) any {
		if cb != nil {
			cb(status, a[0].String())
		}
		onText.Release()
		return nil
	})
	onResp = js.FuncOf(func(_ js.Value, a []js.Value) any {
		status = a[0].Get("status").Int()
		a[0].Call("text").Call("then", onText)
		onResp.Release()
		return nil
	})
	onErr = js.FuncOf(func(_ js.Value, a []js.Value) any {
		if cb != nil {
			cb(0, "")
		}
		onErr.Release()
		return nil
	})
	js.Fetch(url, o).Call("then", onResp).Call("catch", onErr)
}

// RequestBytes performs an uncached fetch like Request but delivers the raw
// response bytes via arrayBuffer, so binary payloads (images, chunks,
// downloads) survive intact; Request's text decoding would corrupt them.
func RequestBytes(url string, opts RequestOptions, cb func(status int, body []byte)) {
	o := js.Object().New()
	if opts.Method != "" {
		o.Set("method", opts.Method)
	}
	if len(opts.Headers) > 0 {
		h := js.Object().New()
		for k, v := range opts.Headers {
			h.Set(k, v)
		}
		o.Set("headers", h)
	}
	if opts.Body != "" {
		o.Set("body", opts.Body)
	}

	status := 0
	var onResp, onBuf, onErr js.Func
	release := func() {
		onResp.Release()
		onBuf.Release()
		onErr.Release()
	}
	onBuf = js.FuncOf(func(_ js.Value, a []js.Value) any {
		u8 := js.Uint8Array().New(a[0])
		body := make([]byte, u8.Get("length").Int())
		js.CopyBytesToGo(body, u8)
		release()
		if cb != nil {
			cb(status, body)
		}
		return nil
	})
	onResp = js.FuncOf(func(_ js.Value, a []js.Value) any {
		status = a[0].Get("status").Int()
		a[0].Call("arrayBuffer").Call("then", onBuf)
		return nil
	})
	onErr = js.FuncOf(func(_ js.Value, a []js.Value) any {
		release()
		if cb != nil {
			cb(0, nil)
		}
		return nil
	})
	js.Global().Call("fetch", url, o).Call("then", onResp).Call("catch", onErr)
}
