//go:build js && wasm

package http

import (
	syscalljs "syscall/js"

	"github.com/rfwlab/rfw/v2/js"
)

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
// invoked on the JS event loop; it may be nil. Browser, network and body-read
// failures complete with status 0 and an empty body. Requests are not retried.
func Request(url string, opts RequestOptions, cb func(status int, body string)) {
	status := 0
	var onResp, onText, onErr js.Func
	done := false
	finish := func(code int, body string) {
		if done {
			return
		}
		done = true
		onResp.Release()
		onText.Release()
		onErr.Release()
		if cb != nil {
			cb(code, body)
		}
	}
	onText = js.SafeFuncOf(func(_ js.Value, a []js.Value) any {
		finish(status, a[0].String())
		return nil
	})
	onResp = js.SafeFuncOf(func(_ js.Value, a []js.Value) any {
		if !requestCall(func() {
			status = a[0].Get("status").Int()
			a[0].Call("text").Call("then", onText).Call("catch", onErr)
		}) {
			finish(0, "")
		}
		return nil
	})
	onErr = js.SafeFuncOf(func(_ js.Value, _ []js.Value) any {
		finish(0, "")
		return nil
	})
	if !requestCall(func() { js.Fetch(url, opts.apply()).Call("then", onResp).Call("catch", onErr) }) {
		finish(0, "")
	}
}

// RequestBytes performs an uncached fetch like Request but delivers the raw
// response bytes via arrayBuffer, so binary payloads (images, chunks,
// downloads) survive intact; Request's text decoding would corrupt them.
func RequestBytes(url string, opts RequestOptions, cb func(status int, body []byte)) {
	status := 0
	var onResp, onBuf, onErr js.Func
	done := false
	finish := func(code int, body []byte) {
		if done {
			return
		}
		done = true
		onResp.Release()
		onBuf.Release()
		onErr.Release()
		if cb != nil {
			cb(code, body)
		}
	}
	onBuf = js.SafeFuncOf(func(_ js.Value, a []js.Value) any {
		var body []byte
		if !requestCall(func() {
			u8 := js.Uint8Array().New(a[0])
			body = make([]byte, u8.Get("length").Int())
			js.CopyBytesToGo(body, u8)
		}) {
			finish(0, nil)
			return nil
		}
		finish(status, body)
		return nil
	})
	onResp = js.SafeFuncOf(func(_ js.Value, a []js.Value) any {
		if !requestCall(func() {
			status = a[0].Get("status").Int()
			a[0].Call("arrayBuffer").Call("then", onBuf).Call("catch", onErr)
		}) {
			finish(0, nil)
		}
		return nil
	})
	onErr = js.SafeFuncOf(func(_ js.Value, _ []js.Value) any {
		finish(0, nil)
		return nil
	})
	if !requestCall(func() { js.Fetch(url, opts.apply()).Call("then", onResp).Call("catch", onErr) }) {
		finish(0, nil)
	}
}

// Browser calls can throw before returning a promise, including during unload.
// Caller panics stay outside this guard and retain the normal error reporting.
func requestCall(call func()) (ok bool) {
	defer func() {
		if err := recover(); err != nil {
			if _, browserError := err.(syscalljs.Error); !browserError {
				panic(err)
			}
			ok = false
		}
	}()
	call()
	return true
}
