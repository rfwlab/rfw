//go:build js && wasm

package http

import (
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/js"
)

func mockFetchResponse(t *testing.T, status int, body string) *int {
	t.Helper()
	previous := js.Get("fetch")
	hits := 0
	fetch := js.FuncOf(func(js.Value, []js.Value) any {
		hits++
		response := js.Get("Response").New(body, map[string]any{"status": status})
		return js.Get("Promise").Call("resolve", response)
	})
	js.Set("fetch", fetch)
	t.Cleanup(func() {
		js.Set("fetch", previous)
		fetch.Release()
	})
	return &hits
}

func settleFetch(t *testing.T, fetch func() error) error {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		err := fetch()
		if err != ErrPending {
			return err
		}
		if time.Now().After(deadline) {
			t.Fatal("fetch remained pending")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestFetchJSONRejectsHTTPErrorWithoutDecoding(t *testing.T) {
	hits := mockFetchResponse(t, 403, `{"denied":true}`)
	const url = "/json-error"
	t.Cleanup(func() { ClearCache(url) })
	starts, completes, status := 0, 0, 0
	RegisterHTTPHook(func(start bool, _ string, code int, _ time.Duration) {
		if start {
			starts++
		} else {
			completes++
			status = code
		}
	})
	t.Cleanup(func() { RegisterHTTPHook(nil) })
	out := map[string]any{"unchanged": true}
	err := settleFetch(t, func() error { return FetchJSON(url, &out) })
	if err == nil || err.Error() != `{"denied":true}` {
		t.Fatalf("HTTP error = %v", err)
	}
	if len(out) != 1 || out["unchanged"] != true {
		t.Fatalf("failed response decoded into target: %v", out)
	}
	if err = FetchJSON(url, &out); err == nil || err == ErrPending || *hits != 1 {
		t.Fatalf("cached error = %v, requests = %d", err, *hits)
	}
	if starts != 1 || completes != 1 || status != 403 {
		t.Fatalf("observer starts = %d, completes = %d, status = %d", starts, completes, status)
	}
	ClearCache(url)
	if err = settleFetch(t, func() error { return FetchJSON(url, &out) }); err == nil || *hits != 2 {
		t.Fatalf("refetch error = %v, requests = %d", err, *hits)
	}
}

func TestFetchJSONRejectsNonJSONHTTPError(t *testing.T) {
	mockFetchResponse(t, 500, "<html>unavailable</html>")
	const url = "/html-error"
	t.Cleanup(func() { ClearCache(url) })
	var out any
	err := settleFetch(t, func() error { return FetchJSON(url, &out) })
	if err == nil || err.Error() != "<html>unavailable</html>" {
		t.Fatalf("HTTP error = %v", err)
	}
}

func TestFetchJSONRejectsEmptyHTTPError(t *testing.T) {
	mockFetchResponse(t, 404, "")
	const url = "/empty-error"
	t.Cleanup(func() { ClearCache(url) })
	var out any
	if err := settleFetch(t, func() error { return FetchJSON(url, &out) }); err == nil || err.Error() != "" {
		t.Fatalf("empty HTTP error = %v", err)
	}
}

func TestFetchTextRejectsHTTPError(t *testing.T) {
	hits := mockFetchResponse(t, 404, "missing")
	const url = "/text-error"
	t.Cleanup(func() { ClearCache(url) })
	var text string
	err := settleFetch(t, func() error {
		var err error
		text, err = FetchText(url)
		return err
	})
	if err == nil || err.Error() != "missing" || text != "" {
		t.Fatalf("text = %q, HTTP error = %v", text, err)
	}
	if _, err = FetchText(url); err == nil || err == ErrPending || *hits != 1 {
		t.Fatalf("cached error = %v, requests = %d", err, *hits)
	}
}

func TestFetchSuccessfulResponsesRemainCached(t *testing.T) {
	hits := mockFetchResponse(t, 200, `{"value":7}`)
	const url = "/success"
	t.Cleanup(func() { ClearCache(url) })
	var out struct{ Value int }
	if err := settleFetch(t, func() error { return FetchJSON(url, &out) }); err != nil || out.Value != 7 {
		t.Fatalf("JSON = %v, error = %v", out, err)
	}
	if err := FetchJSON(url, &out); err != nil || *hits != 1 {
		t.Fatalf("cached JSON error = %v, requests = %d", err, *hits)
	}
	var text string
	err := settleFetch(t, func() error {
		var err error
		text, err = FetchText(url)
		return err
	})
	if err != nil || text != `{"value":7}` {
		t.Fatalf("text = %q, error = %v", text, err)
	}
	if _, err = FetchText(url); err != nil || *hits != 2 {
		t.Fatalf("cached text error = %v, requests = %d", err, *hits)
	}
}

func TestFetchNetworkFailureSettlesBothCaches(t *testing.T) {
	previous := js.Get("fetch")
	fetch := js.FuncOf(func(js.Value, []js.Value) any {
		return js.Get("Promise").Call("reject", js.Get("Error").New("offline"))
	})
	js.Set("fetch", fetch)
	t.Cleanup(func() {
		js.Set("fetch", previous)
		fetch.Release()
		ClearCache("/offline")
	})
	var out any
	if err := settleFetch(t, func() error { return FetchJSON("/offline", &out) }); err == nil || out != nil {
		t.Fatalf("network JSON = %v, error = %v", out, err)
	}
	if err := settleFetch(t, func() error {
		_, err := FetchText("/offline")
		return err
	}); err == nil {
		t.Fatal("network text returned success")
	}
}
