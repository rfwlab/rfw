//go:build js && wasm

// Package http provides cached browser fetch helpers.
package http

import (
	"encoding/json"
	"errors"
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

// SetNativeClient has no effect in browser builds. It takes any so the
// browser build does not import net/http, which would link the whole Go
// HTTP and TLS stack into every bundle. Native builds keep the typed
// *http.Client signature.
func SetNativeClient(_ any) {}

// FetchJSON retrieves JSON data from the given URL and decodes it into v.
// Results are cached by URL. If a request is already in progress, FetchJSON
// returns ErrPending. HTTP status codes >= 400 return an error containing the
// response text without decoding into v. ClearCache allows retrying a failure.
func FetchJSON(url string, v any) error {
	ceIface, _ := cache.LoadOrStore(url, &cacheEntry{ready: make(chan struct{})})
	ce := ceIface.(*cacheEntry)

	ce.once.Do(func() {
		go func() {
			hook := currentHTTPHook()
			notifyHTTPHook(hook, true, url, 0, 0)
			start := time.Now()
			Request(url, RequestOptions{}, func(status int, body string) {
				switch {
				case status == 0:
					ce.err = errors.New("http: request failed")
				case status >= 400:
					ce.err = errors.New(body)
				case !json.Valid([]byte(body)):
					ce.err = errors.New("http: invalid JSON response")
				default:
					ce.data = []byte(body)
				}
				notifyHTTPHook(hook, false, url, status, time.Since(start))
				close(ce.ready)
			})
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
// HTTP status codes >= 400 return an error containing the response text.
func FetchText(url string) (string, error) {
	ceIface, _ := textCache.LoadOrStore(url, &textEntry{ready: make(chan struct{})})
	ce := ceIface.(*textEntry)

	ce.once.Do(func() {
		go func() {
			hook := currentHTTPHook()
			notifyHTTPHook(hook, true, url, 0, 0)
			start := time.Now()
			Request(url, RequestOptions{}, func(status int, body string) {
				switch {
				case status == 0:
					ce.err = errors.New("http: request failed")
				case status >= 400:
					ce.err = errors.New(body)
				default:
					ce.text = body
				}
				notifyHTTPHook(hook, false, url, status, time.Since(start))
				close(ce.ready)
			})
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
