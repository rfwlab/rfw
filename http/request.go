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
	// BodyValue is a browser-side body that is not a string: FormData for
	// multipart uploads, Blob, ArrayBuffer. It takes precedence over Body.
	BodyValue js.Value
}

// apply builds the fetch init object for the request.
func (o RequestOptions) apply() js.Value {
	init := js.Object().New()
	if o.Method != "" {
		init.Set("method", o.Method)
	}
	if len(o.Headers) > 0 {
		h := js.Object().New()
		for k, v := range o.Headers {
			h.Set(k, v)
		}
		init.Set("headers", h)
	}
	if o.BodyValue.Truthy() {
		init.Set("body", o.BodyValue)
	} else if o.Body != "" {
		init.Set("body", o.Body)
	}
	return init
}

// Request performs an uncached fetch with a custom method, headers and body and
// invokes cb with the HTTP status and the response body text once it resolves.
//
// Unlike FetchJSON/FetchText it does not cache and carries request headers, so
// it is the right primitive for authenticated and mutating requests (the caller
// supplies Authorization / workspace headers via RequestOptions.Headers). cb is
// invoked on the JS event loop; it may be nil.
func Request(url string, opts RequestOptions, cb func(status int, body string)) {
	status := 0
	var onResp, onText, onErr js.Func
	// Every callback is released on both outcomes: releasing only the one that
	// fired leaks the other two for the lifetime of the page.
	release := func() {
		onResp.Release()
		onText.Release()
		onErr.Release()
	}
	onText = js.SafeFuncOf(func(_ js.Value, a []js.Value) any {
		release()
		if cb != nil {
			cb(status, a[0].String())
		}
		return nil
	})
	onResp = js.SafeFuncOf(func(_ js.Value, a []js.Value) any {
		status = a[0].Get("status").Int()
		a[0].Call("text").Call("then", onText).Call("catch", onErr)
		return nil
	})
	onErr = js.SafeFuncOf(func(_ js.Value, _ []js.Value) any {
		release()
		if cb != nil {
			cb(0, "")
		}
		return nil
	})
	js.Fetch(url, opts.apply()).Call("then", onResp).Call("catch", onErr)
}

// RequestBytes performs an uncached fetch like Request but delivers the raw
// response bytes via arrayBuffer, so binary payloads (images, chunks,
// downloads) survive intact; Request's text decoding would corrupt them.
func RequestBytes(url string, opts RequestOptions, cb func(status int, body []byte)) {
	status := 0
	var onResp, onBuf, onErr js.Func
	release := func() {
		onResp.Release()
		onBuf.Release()
		onErr.Release()
	}
	onBuf = js.SafeFuncOf(func(_ js.Value, a []js.Value) any {
		u8 := js.Uint8Array().New(a[0])
		body := make([]byte, u8.Get("length").Int())
		js.CopyBytesToGo(body, u8)
		release()
		if cb != nil {
			cb(status, body)
		}
		return nil
	})
	onResp = js.SafeFuncOf(func(_ js.Value, a []js.Value) any {
		status = a[0].Get("status").Int()
		a[0].Call("arrayBuffer").Call("then", onBuf).Call("catch", onErr)
		return nil
	})
	onErr = js.SafeFuncOf(func(_ js.Value, _ []js.Value) any {
		release()
		if cb != nil {
			cb(0, nil)
		}
		return nil
	})
	js.Fetch(url, opts.apply()).Call("then", onResp).Call("catch", onErr)
}
