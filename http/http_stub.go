//go:build !js || !wasm

package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"
	"sync"
	"time"

	"github.com/rfwlab/rfw/v2/internal/safehttp"
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
var httpHookMu sync.RWMutex
var httpClientMu sync.RWMutex
var httpClient = safehttp.NewClient()

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

// SetNativeClient replaces the native HTTP client. Custom clients may reach
// private networks and must only receive trusted URLs. Passing nil restores
// the default client, which rejects private network addresses.
func SetNativeClient(client *stdhttp.Client) {
	if client == nil {
		client = safehttp.NewClient()
	}
	httpClientMu.Lock()
	httpClient = client
	httpClientMu.Unlock()
}

func currentNativeClient() *stdhttp.Client {
	httpClientMu.RLock()
	client := httpClient
	httpClientMu.RUnlock()
	return client
}

func fetchBytes(rawURL string) (status int, body []byte, err error) {
	req, err := safehttp.NewRequest(context.Background(), stdhttp.MethodGet, rawURL)
	if err != nil {
		return 0, nil, err
	}
	resp, err := currentNativeClient().Do(req)
	if err != nil {
		return 0, nil, err
	}

	b, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if closeErr != nil {
		return resp.StatusCode, nil, closeErr
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
			hook := currentHTTPHook()
			if hook != nil {
				hook(true, url, 0, 0)
			}
			start := time.Now()
			status, b, err := fetchBytes(url)
			ce.data = b
			ce.err = err
			if hook != nil {
				hook(false, url, status, time.Since(start))
			}
			close(ce.ready)
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
			if hook != nil {
				hook(true, url, 0, 0)
			}
			start := time.Now()
			status, b, err := fetchBytes(url)
			ce.text = string(b)
			ce.err = err
			if hook != nil {
				hook(false, url, status, time.Since(start))
			}
			close(ce.ready)
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
