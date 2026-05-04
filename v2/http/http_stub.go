//go:build !js || !wasm

package http

import (
"encoding/json"
"errors"
"io"
stdhttp "net/http"
"sync"
"time"
)

// ErrPending is returned when a fetch request is still in flight.
var ErrPending = errors.New("http: request pending")

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

func fetchBytes(url string) (status int, body []byte, err error) {
resp, err := stdhttp.Get(url)
if err != nil {
return 0, nil, err
}
defer resp.Body.Close()

b, err := io.ReadAll(resp.Body)
if err != nil {
return resp.StatusCode, nil, err
}
if resp.StatusCode >= 400 {
return resp.StatusCode, nil, errors.New(string(b))
}
return resp.StatusCode, b, nil
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
status, b, err := fetchBytes(url)
ce.data = b
ce.err = err
close(ce.ready)
if httpHook != nil {
httpHook(false, url, status, time.Since(start))
}
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
status, b, err := fetchBytes(url)
ce.text = string(b)
ce.err = err
close(ce.ready)
if httpHook != nil {
httpHook(false, url, status, time.Since(start))
}
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
