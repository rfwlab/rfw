//go:build js && wasm

package http

import (
	"context"
	"sync"
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
	RequestContext(context.Background(), url, opts, cb)
}

// RequestContext is Request with browser fetch/body cancellation. Cancelled or
// expired contexts complete once with status0 and an empty body. Abort does not
// prove that a server mutation rolled back; callers must reconcile unknown writes.
// A nil context has the same unbounded lifetime as Request.
func RequestContext(ctx context.Context, url string, opts RequestOptions, cb func(status int, body string)) {
	requestBody(ctx, url, opts, "text", func(status int, value js.Value) {
		if cb == nil {
			return
		}
		if status == 0 {
			cb(0, "")
			return
		}
		cb(status, value.String())
	})
}

// RequestBytes performs an uncached fetch like Request but delivers the raw
// response bytes via arrayBuffer, so binary payloads survive text decoding.
func RequestBytes(url string, opts RequestOptions, cb func(status int, body []byte)) {
	RequestBytesContext(context.Background(), url, opts, cb)
}

// RequestBytesContext is RequestBytes with the cancellation contract of
// RequestContext. Cancellation also aborts a pending response-body read.
func RequestBytesContext(ctx context.Context, url string, opts RequestOptions, cb func(status int, body []byte)) {
	requestBody(ctx, url, opts, "arrayBuffer", func(status int, value js.Value) {
		if cb == nil {
			return
		}
		if status == 0 {
			cb(0, nil)
			return
		}
		var body []byte
		if !requestCall(func() {
			bytes := js.Uint8Array().New(value)
			body = make([]byte, bytes.Get("length").Int())
			js.CopyBytesToGo(body, bytes)
		}) {
			cb(0, nil)
			return
		}
		cb(status, body)
	})
}

func requestBody(ctx context.Context, url string, opts RequestOptions, read string, cb func(int, js.Value)) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		cb(0, js.Undefined())
		return
	}
	var controller js.Value
	if ctx.Done() != nil {
		constructor := js.Get("AbortController")
		if constructor.Type() != js.TypeFunction || !requestCall(func() { controller = constructor.New() }) {
			cb(0, js.Undefined())
			return
		}
	}
	status := 0
	var onResp, onBody, onErr js.Func
	var stop func() bool
	var once sync.Once
	finish := func(code int, body js.Value) {
		once.Do(func() {
			if stop != nil {
				stop()
			}
			onResp.Release()
			onBody.Release()
			onErr.Release()
			if ctx.Err() != nil {
				code, body = 0, js.Undefined()
			}
			cb(code, body)
		})
	}
	// one promise chain reaches either onBody or onErr. Cancellation aborts
	// fetch/body I/O; release callbacks only after that chain has settled.
	onResp = js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
		var body js.Value
		if !requestCall(func() {
			status = args[0].Get("status").Int()
			body = args[0].Call(read)
		}) {
			return js.Get("Promise").Call("reject", js.Undefined())
		}
		return body
	})
	onBody = js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
		finish(status, args[0])
		return nil
	})
	onErr = js.SafeFuncOf(func(_ js.Value, _ []js.Value) any {
		finish(0, js.Undefined())
		return nil
	})
	if controller.Truthy() {
		stop = context.AfterFunc(ctx, func() { requestCall(func() { controller.Call("abort") }) })
	}
	if !requestCall(func() {
		init := opts.apply()
		if controller.Truthy() {
			init.Set("signal", controller.Get("signal"))
		}
		js.Fetch(url, init).Call("then", onResp).Call("then", onBody, onErr)
	}) {
		finish(0, js.Undefined())
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
