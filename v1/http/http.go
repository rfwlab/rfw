//go:build js && wasm

package http

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/rfwlab/rfw/v1/js"
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

// RegisterHTTPHook adds a callback invoked on request start and completion.
// The callback receives a start flag, request URL, status code and duration.
var httpHook func(start bool, url string, status int, duration time.Duration)

// RegisterHTTPHook registers fn to receive HTTP request events.
func RegisterHTTPHook(fn func(start bool, url string, status int, duration time.Duration)) {
	httpHook = fn
}

// FetchJSON retrieves JSON data from the given URL and decodes it into v.
// Results are cached by URL. If a request is already in progress, FetchJSON
// returns ErrPending.
func FetchJSON(url string, v any) error {
	ceIface, _ := cache.LoadOrStore(url, &cacheEntry{ready: make(chan struct{})})
	ce := ceIface.(*cacheEntry)

	ce.once.Do(func() {
		go func() {
			if httpHook != nil {
				httpHook(true, url, 0, 0)
			}
			start := time.Now()
			js.Fetch(url).Call("then",
				js.FuncOf(func(this js.Value, args []js.Value) any {
					resp := args[0]
					status := resp.Get("status").Int()
					resp.Call("json").Call("then",
						js.FuncOf(func(this js.Value, args []js.Value) any {
							obj := args[0]
							jsonStr := js.JSON().Call("stringify", obj).String()
							ce.data = []byte(jsonStr)
							close(ce.ready)
							if httpHook != nil {
								httpHook(false, url, status, time.Since(start))
							}
							return nil
						}),
						js.FuncOf(func(this js.Value, args []js.Value) any {
							ce.err = errors.New(args[0].String())
							close(ce.ready)
							if httpHook != nil {
								httpHook(false, url, status, time.Since(start))
							}
							return nil
						}),
					)
					return nil
				}),
				js.FuncOf(func(this js.Value, args []js.Value) any {
					ce.err = errors.New(args[0].String())
					close(ce.ready)
					if httpHook != nil {
						httpHook(false, url, 0, time.Since(start))
					}
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
			if httpHook != nil {
				httpHook(true, url, 0, 0)
			}
			start := time.Now()
			js.Fetch(url).Call("then",
				js.FuncOf(func(this js.Value, args []js.Value) any {
					resp := args[0]
					status := resp.Get("status").Int()
					resp.Call("text").Call("then",
						js.FuncOf(func(this js.Value, args []js.Value) any {
							ce.text = args[0].String()
							close(ce.ready)
							if httpHook != nil {
								httpHook(false, url, status, time.Since(start))
							}
							return nil
						}),
						js.FuncOf(func(this js.Value, args []js.Value) any {
							ce.err = errors.New(args[0].String())
							close(ce.ready)
							if httpHook != nil {
								httpHook(false, url, status, time.Since(start))
							}
							return nil
						}),
					)
					return nil
				}),
				js.FuncOf(func(this js.Value, args []js.Value) any {
					ce.err = errors.New(args[0].String())
					close(ce.ready)
					if httpHook != nil {
						httpHook(false, url, 0, time.Since(start))
					}
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
