//go:build js && wasm

package http

import (
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/js"
)

func requestFetch(t *testing.T, body string) {
	t.Helper()
	previous := js.Get("fetch")
	js.Set("fetch", js.Get("Function").New(body))
	t.Cleanup(func() { js.Set("fetch", previous) })
}

func waitRequest(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("request did not complete")
	}
	time.Sleep(time.Millisecond)
}

func TestRequestCompletesWhenFetchThrows(t *testing.T) {
	requestFetch(t, `throw new TypeError("fetch refused")`)
	calls, status, body := 0, -1, "unset"
	Request("/sync-error", RequestOptions{Method: "POST"}, func(code int, value string) { calls++; status, body = code, value })
	if calls != 1 || status != 0 || body != "" {
		t.Fatalf("completion = %d, %d, %q", calls, status, body)
	}
}

func TestRequestBytesCompletesWhenFetchThrows(t *testing.T) {
	requestFetch(t, `throw new TypeError("fetch refused")`)
	calls, status := 0, -1
	body := []byte{1}
	RequestBytes("/sync-error", RequestOptions{}, func(code int, value []byte) { calls++; status, body = code, value })
	if calls != 1 || status != 0 || body != nil {
		t.Fatalf("completion = %d, %d, %v", calls, status, body)
	}
}

func TestRequestCompletesWhenTextThrows(t *testing.T) {
	requestFetch(t, `return Promise.resolve({status:200,text(){throw new TypeError("text refused")}})`)
	done := make(chan struct{}, 2)
	calls, status, body := 0, -1, "unset"
	Request("/text-error", RequestOptions{}, func(code int, value string) { calls++; status, body = code, value; done <- struct{}{} })
	waitRequest(t, done)
	if calls != 1 || status != 0 || body != "" {
		t.Fatalf("completion = %d, %d, %q", calls, status, body)
	}
}

func TestRequestBytesCompletesWhenArrayBufferThrows(t *testing.T) {
	requestFetch(t, `return Promise.resolve({status:200,arrayBuffer(){throw new TypeError("buffer refused")}})`)
	done := make(chan struct{}, 2)
	calls, status := 0, -1
	body := []byte{1}
	RequestBytes("/buffer-error", RequestOptions{}, func(code int, value []byte) { calls++; status, body = code, value; done <- struct{}{} })
	waitRequest(t, done)
	if calls != 1 || status != 0 || body != nil {
		t.Fatalf("completion = %d, %d, %v", calls, status, body)
	}
}

func TestRequestCompletesWhenBodyPromiseRejects(t *testing.T) {
	requestFetch(t, `return Promise.resolve({status:200,text(){return Promise.reject(new TypeError("read failed"))}})`)
	done := make(chan struct{}, 2)
	calls, status, body := 0, -1, "unset"
	Request("/body-error", RequestOptions{}, func(code int, value string) { calls++; status, body = code, value; done <- struct{}{} })
	waitRequest(t, done)
	if calls != 1 || status != 0 || body != "" {
		t.Fatalf("completion = %d, %d, %q", calls, status, body)
	}
}

func TestRequestBytesCompletesWhenBodyPromiseRejects(t *testing.T) {
	requestFetch(t, `return Promise.resolve({status:200,arrayBuffer(){return Promise.reject(new TypeError("read failed"))}})`)
	done := make(chan struct{}, 2)
	calls, status := 0, -1
	body := []byte{1}
	RequestBytes("/body-error", RequestOptions{}, func(code int, value []byte) { calls++; status, body = code, value; done <- struct{}{} })
	waitRequest(t, done)
	if calls != 1 || status != 0 || body != nil {
		t.Fatalf("completion = %d, %d, %v", calls, status, body)
	}
}

func TestRequestKeepsHTTPStatusAndText(t *testing.T) {
	requestFetch(t, `return Promise.resolve(new Response("missing",{status:404}))`)
	done := make(chan struct{}, 2)
	calls, status, body := 0, -1, "unset"
	Request("/missing", RequestOptions{}, func(code int, value string) { calls++; status, body = code, value; done <- struct{}{} })
	waitRequest(t, done)
	if calls != 1 || status != 404 || body != "missing" {
		t.Fatalf("completion = %d, %d, %q", calls, status, body)
	}
}

func TestRequestBytesKeepsBinaryBody(t *testing.T) {
	requestFetch(t, `return Promise.resolve(new Response(new Uint8Array([0,255,128,1]),{status:201}))`)
	done := make(chan struct{}, 2)
	calls, status := 0, -1
	var body []byte
	RequestBytes("/binary", RequestOptions{}, func(code int, value []byte) { calls++; status, body = code, value; done <- struct{}{} })
	waitRequest(t, done)
	if calls != 1 || status != 201 || len(body) != 4 || body[0] != 0 || body[1] != 255 || body[2] != 128 || body[3] != 1 {
		t.Fatalf("completion = %d, %d, %v", calls, status, body)
	}
}

func TestRequestFailureAllowsNilCallback(t *testing.T) {
	requestFetch(t, `throw new TypeError("fetch refused")`)
	Request("/nil", RequestOptions{}, nil)
	RequestBytes("/nil", RequestOptions{}, nil)
}

func TestRequestDoesNotConvertCallerPanicToTransportFailure(t *testing.T) {
	requestFetch(t, `throw new TypeError("fetch refused")`)
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		Request("/callback", RequestOptions{}, func(int, string) { panic("caller failure") })
	}()
	if recovered != "caller failure" {
		t.Fatalf("caller panic = %v", recovered)
	}
}

func TestFetchJSONCompletesWhenFetchThrows(t *testing.T) {
	requestFetch(t, `throw new TypeError("fetch refused")`)
	const url = "/cached-sync-json"
	t.Cleanup(func() { ClearCache(url) })
	var value map[string]any
	if err := settleFetch(t, func() error { return FetchJSON(url, &value) }); err == nil || value != nil {
		t.Fatalf("JSON value = %v, error = %v", value, err)
	}
}

func TestFetchTextCompletesWhenFetchThrows(t *testing.T) {
	requestFetch(t, `throw new TypeError("fetch refused")`)
	const url = "/cached-sync-text"
	t.Cleanup(func() { ClearCache(url) })
	var value string
	err := settleFetch(t, func() error { var err error; value, err = FetchText(url); return err })
	if err == nil || value != "" {
		t.Fatalf("text value = %q, error = %v", value, err)
	}
}

func TestFetchJSONDoesNotRoundIntegerThroughJavaScript(t *testing.T) {
	requestFetch(t, `return Promise.resolve(new Response('{"id":9007199254740993}',{status:200}))`)
	const url = "/json-integer"
	t.Cleanup(func() { ClearCache(url) })
	var value struct {
		ID int64 `json:"id"`
	}
	if err := settleFetch(t, func() error { return FetchJSON(url, &value) }); err != nil || value.ID != 9007199254740993 {
		t.Fatalf("JSON value = %v, error = %v", value, err)
	}
}

func TestRequestReportsCallerJavaScriptPanicAfterCompletion(t *testing.T) {
	requestFetch(t, `return Promise.resolve(new Response("ok",{status:200}))`)
	previous := js.OnFuncPanic
	t.Cleanup(func() { js.OnFuncPanic = previous })
	done := make(chan struct{}, 2)
	var reported any
	js.OnFuncPanic = func(value any, _ []byte) { reported = value; done <- struct{}{} }
	calls := 0
	fail := js.Get("Function").New(`throw new TypeError("caller JavaScript failure")`)
	Request("/caller-js", RequestOptions{}, func(status int, _ string) {
		calls++
		if status != 200 {
			t.Fatalf("callback status = %d", status)
		}
		fail.Invoke()
	})
	waitRequest(t, done)
	if calls != 1 || reported == nil {
		t.Fatalf("completion calls = %d, reported = %v", calls, reported)
	}
}
