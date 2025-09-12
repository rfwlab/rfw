//go:build js && wasm

package assets

import (
	"errors"
	"sync"

	"github.com/rfwlab/rfw/v1/http"
	"github.com/rfwlab/rfw/v1/js"
)

// loadImageFn loads an image and invokes done on completion.
var loadImageFn = func(url string, done func(js.Value, error)) {
	img := js.Image().New()
	onload := js.FuncOf(func(this js.Value, args []js.Value) any {
		done(img, nil)
		return nil
	})
	onerror := js.FuncOf(func(this js.Value, args []js.Value) any {
		done(js.Value{}, errors.New("assets: image load failed"))
		return nil
	})
	img.Set("onload", onload)
	img.Set("onerror", onerror)
	img.Set("src", url)
}

// loadBinaryFn fetches binary data and invokes done on completion.
var loadBinaryFn = func(url string, done func([]byte, error)) {
	js.Fetch(url).Call("then",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			resp := args[0]
			resp.Call("arrayBuffer").Call("then",
				js.FuncOf(func(this js.Value, args []js.Value) any {
					buf := js.Uint8Array().New(args[0])
					length := buf.Get("length").Int()
					data := make([]byte, length)
					for i := 0; i < length; i++ {
						data[i] = byte(buf.Index(i).Int())
					}
					done(data, nil)
					return nil
				}),
				js.FuncOf(func(this js.Value, args []js.Value) any {
					done(nil, errors.New(args[0].String()))
					return nil
				}),
			)
			return nil
		}),
		js.FuncOf(func(this js.Value, args []js.Value) any {
			done(nil, errors.New(args[0].String()))
			return nil
		}),
	)
}

// imageEntry holds the result of an image load.
type imageEntry struct {
	once  sync.Once
	img   js.Value
	err   error
	ready chan struct{}
}

var imageCache sync.Map // map[string]*imageEntry

// LoadImage asynchronously loads an image from url.
// While loading it returns http.ErrPending.
// Results are cached by URL.
func LoadImage(url string) (js.Value, error) {
	ceIface, _ := imageCache.LoadOrStore(url, &imageEntry{ready: make(chan struct{})})
	ce := ceIface.(*imageEntry)

	ce.once.Do(func() {
		go loadImageFn(url, func(v js.Value, err error) {
			ce.img = v
			ce.err = err
			close(ce.ready)
		})
	})

	select {
	case <-ce.ready:
		if ce.err != nil {
			return js.Value{}, ce.err
		}
		return ce.img, nil
	default:
		return js.Value{}, http.ErrPending
	}
}

// modelEntry holds the result of a binary load.
type modelEntry struct {
	once  sync.Once
	data  []byte
	err   error
	ready chan struct{}
}

var modelCache sync.Map // map[string]*modelEntry

// LoadModel fetches binary data from url. It caches results and
// returns http.ErrPending while the request is in flight.
func LoadModel(url string) ([]byte, error) {
	ceIface, _ := modelCache.LoadOrStore(url, &modelEntry{ready: make(chan struct{})})
	ce := ceIface.(*modelEntry)

	ce.once.Do(func() {
		go loadBinaryFn(url, func(b []byte, err error) {
			ce.data = b
			ce.err = err
			close(ce.ready)
		})
	})

	select {
	case <-ce.ready:
		if ce.err != nil {
			return nil, ce.err
		}
		return ce.data, nil
	default:
		return nil, http.ErrPending
	}
}

// LoadJSON delegates to http.FetchJSON and shares its caching behavior.
func LoadJSON(url string, v any) error { return http.FetchJSON(url, v) }

// ClearCache removes cached assets for url.
func ClearCache(url string) {
	imageCache.Delete(url)
	modelCache.Delete(url)
}
