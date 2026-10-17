//go:build js && wasm

package http

import (
	"context"
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/js"
)

func TestRequestContextCancelledBeforeFetch(t *testing.T) {
	requestFetch(t, `throw new Error("cancelled request reached fetch")`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls, status, body := 0, -1, "unset"
	RequestContext(ctx, "/cancelled", RequestOptions{}, func(code int, value string) { calls++; status, body = code, value })
	if calls != 1 || status != 0 || body != "" {
		t.Fatalf("completion = %d, %d, %q", calls, status, body)
	}
}

func TestRequestContextCancellationAbortsFetchOnce(t *testing.T) {
	requestFetch(t, `const signal=arguments[1].signal; globalThis.__requestSignal=signal;
 return new Promise((resolve,reject)=>signal.addEventListener('abort',()=>reject(new DOMException('aborted','AbortError')),{once:true}))`)
	t.Cleanup(func() { js.Global().Delete("__requestSignal") })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{}, 2)
	calls, status, body := 0, -1, "unset"
	RequestContext(ctx, "/pending", RequestOptions{Method: "POST", Body: "one"}, func(code int, value string) { calls++; status, body = code, value; done <- struct{}{} })
	cancel()
	waitRequest(t, done)
	if calls != 1 || status != 0 || body != "" || !js.Get("__requestSignal").Get("aborted").Bool() {
		t.Fatalf("completion = %d, %d, %q", calls, status, body)
	}
}

func TestRequestContextDeadlineAbortsPendingFetch(t *testing.T) {
	requestFetch(t, `const signal=arguments[1].signal;
 return new Promise((resolve,reject)=>signal.addEventListener('abort',()=>reject(new DOMException('aborted','AbortError')),{once:true}))`)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	done := make(chan struct{}, 2)
	calls, status := 0, -1
	RequestContext(ctx, "/deadline", RequestOptions{}, func(code int, _ string) { calls++; status = code; done <- struct{}{} })
	waitRequest(t, done)
	if calls != 1 || status != 0 || ctx.Err() != context.DeadlineExceeded {
		t.Fatalf("completion = %d, %d, %v", calls, status, ctx.Err())
	}
}

func TestRequestBytesContextCancelsPendingBody(t *testing.T) {
	requestFetch(t, `const signal=arguments[1].signal;
 return Promise.resolve({status:200,arrayBuffer(){globalThis.__requestBodyPending=true;return new Promise((resolve,reject)=>signal.addEventListener('abort',()=>reject(new DOMException('aborted','AbortError')),{once:true}))}})`)
	t.Cleanup(func() { js.Global().Delete("__requestBodyPending") })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{}, 2)
	calls, status := 0, -1
	body := []byte{1}
	RequestBytesContext(ctx, "/pending-body", RequestOptions{}, func(code int, value []byte) { calls++; status, body = code, value; done <- struct{}{} })
	deadline := time.Now().Add(time.Second)
	for !js.Get("__requestBodyPending").Truthy() {
		if time.Now().After(deadline) {
			t.Fatal("body reader did not start")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	waitRequest(t, done)
	if calls != 1 || status != 0 || body != nil {
		t.Fatalf("completion = %d, %d, %v", calls, status, body)
	}
}

func TestRequestContextStopsCancellationAfterSuccess(t *testing.T) {
	requestFetch(t, `globalThis.__requestSignal=arguments[1].signal;return Promise.resolve(new Response('saved',{status:201}))`)
	t.Cleanup(func() { js.Global().Delete("__requestSignal") })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{}, 2)
	calls, status, body := 0, -1, "unset"
	RequestContext(ctx, "/saved", RequestOptions{}, func(code int, value string) { calls++; status, body = code, value; done <- struct{}{} })
	waitRequest(t, done)
	cancel()
	time.Sleep(10 * time.Millisecond)
	if calls != 1 || status != 201 || body != "saved" || js.Get("__requestSignal").Get("aborted").Bool() {
		t.Fatalf("completion = %d, %d, %q", calls, status, body)
	}
}

func TestRequestContextFailsClosedWithoutAbortController(t *testing.T) {
	previous := js.Get("AbortController")
	js.Set("AbortController", js.Undefined())
	t.Cleanup(func() { js.Set("AbortController", previous) })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls, status := 0, -1
	RequestContext(ctx, "/unsupported", RequestOptions{}, func(code int, _ string) { calls++; status = code })
	if calls != 1 || status != 0 {
		t.Fatalf("completion = %d, %d", calls, status)
	}
}
