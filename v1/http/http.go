//go:build js && wasm

package http

import (
	"encoding/json"
	"errors"
	"sync"

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

var cache sync.Map // map[string]*cacheEntry

// FetchJSON retrieves JSON data from the given URL and decodes it into v.
// Results are cached by URL. If a request is already in progress, FetchJSON
// returns ErrPending.
func FetchJSON(url string, v any) error {
	ceIface, _ := cache.LoadOrStore(url, &cacheEntry{ready: make(chan struct{})})
	ce := ceIface.(*cacheEntry)

	ce.once.Do(func() {
		go func() {
			js.Fetch(url).Call("then",
				js.FuncOf(func(this js.Value, args []js.Value) any {
					resp := args[0]
					resp.Call("json").Call("then",
						js.FuncOf(func(this js.Value, args []js.Value) any {
							obj := args[0]
							jsonStr := js.JSON().Call("stringify", obj).String()
							ce.data = []byte(jsonStr)
							close(ce.ready)
							return nil
						}),
						js.FuncOf(func(this js.Value, args []js.Value) any {
							ce.err = errors.New(args[0].String())
							close(ce.ready)
							return nil
						}),
					)
					return nil
				}),
				js.FuncOf(func(this js.Value, args []js.Value) any {
					ce.err = errors.New(args[0].String())
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

// ClearCache removes any cached response for the given URL.
func ClearCache(url string) {
	cache.Delete(url)
}
