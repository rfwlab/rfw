//go:build js && wasm

// Package http provides cached browser fetch helpers.
package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"sync"
	"time"

	"github.com/rfwlab/rfw/v2/js"
)

// ErrPending is returned when a fetch request is still in flight.
var ErrPending = errors.New("http: request pending")

// cacheEntry holds the result of a fetch operation.
type cacheEntry struct {
	once  sync.Once
	data  []byte
	err   error
	ready chan struct{}
}
type textEntry struct {
	once  sync.Once
	text  string
	err   error
	ready chan struct{}
}

var cache sync.Map     // map[string]*cacheEntry
var textCache sync.Map // map[string]*textEntry
var httpHookMu sync.RWMutex

// RegisterHTTPHook adds a callback invoked on request start and completion.
// The callback receives a start flag, request URL, status code and duration.
var httpHook func(start bool, url string, status int, duration time.Duration)

// RegisterHTTPHook registers fn to receive HTTP request events.
func RegisterHTTPHook(fn func(start bool, url string, status int, duration time.Duration)) {
	httpHookMu.Lock()
	httpHook = fn
	httpHookMu.Unlock()
}

func currentHTTPHook() func(bool, string, int, time.Duration) {
	httpHookMu.RLock()
	hook := httpHook
	httpHookMu.RUnlock()
	return hook
}

func notifyHTTPHook(hook func(bool, string, int, time.Duration), start bool, url string, status int, duration time.Duration) {
	if hook == nil {
		return
	}
	js.Guard("HTTP observer", func() { hook(start, url, status, duration) })
}

// SetNativeClient has no effect in browser builds.
func SetNativeClient(_ *stdhttp.Client) {}

// FetchJSON retrieves JSON data from the given URL and decodes it into v.
// Results are cached by URL. If a request is already in progress, FetchJSON
// returns ErrPending.
func FetchJSON(url string, v any) error {
	ceIface, _ := cache.LoadOrStore(url, &cacheEntry{ready: make(chan struct{})})
	ce := ceIface.(*cacheEntry)

	ce.once.Do(func() {
		go func() {
			hook := currentHTTPHook()
			notifyHTTPHook(hook, true, url, 0, 0)
			start := time.Now()
			js.Fetch(url).Call("then",
				js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
					resp := args[0]
					status := resp.Get("status").Int()
					resp.Call("json").Call("then",
						js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
							obj := args[0]
							jsonStr := js.GlobalJSON().Call("stringify", obj).String()
							ce.data = []byte(jsonStr)
							notifyHTTPHook(hook, false, url, status, time.Since(start))
							close(ce.ready)
							return nil
						}),
						js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
							ce.err = errors.New(args[0].String())
							notifyHTTPHook(hook, false, url, status, time.Since(start))
							close(ce.ready)
							return nil
						}),
					)
					return nil
				}),
				js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
					ce.err = errors.New(args[0].String())
					notifyHTTPHook(hook, false, url, 0, time.Since(start))
					close(ce.ready)
					return nil
				}),
			)
		}()
	})

	select {
	case <-ce.ready:
		if ce.err != nil {
			return ce.err
		}
		return json.Unmarshal(ce.data, v)
	default:
		return ErrPending
	}
}

// FetchText retrieves text data from url. Results are cached by URL.
// If a request is already in progress, FetchText returns ErrPending.
func FetchText(url string) (string, error) {
	ceIface, _ := textCache.LoadOrStore(url, &textEntry{ready: make(chan struct{})})
	ce := ceIface.(*textEntry)

	ce.once.Do(func() {
		go func() {
			hook := currentHTTPHook()
			notifyHTTPHook(hook, true, url, 0, 0)
			start := time.Now()
			js.Fetch(url).Call("then",
				js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
					resp := args[0]
					status := resp.Get("status").Int()
					resp.Call("text").Call("then",
						js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
							ce.text = args[0].String()
							notifyHTTPHook(hook, false, url, status, time.Since(start))
							close(ce.ready)
							return nil
						}),
						js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
							ce.err = errors.New(args[0].String())
							notifyHTTPHook(hook, false, url, status, time.Since(start))
							close(ce.ready)
							return nil
						}),
					)
					return nil
				}),
				js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
					ce.err = errors.New(args[0].String())
					notifyHTTPHook(hook, false, url, 0, time.Since(start))
					close(ce.ready)
					return nil
				}),
			)
		}()
	})

	select {
	case <-ce.ready:
		if ce.err != nil {
			return "", ce.err
		}
		return ce.text, nil
	default:
		return "", ErrPending
	}
}

// ClearCache removes any cached response for the given URL.
func ClearCache(url string) {
	cache.Delete(url)
	textCache.Delete(url)
}
