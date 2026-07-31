//go:build !js || !wasm

// Package assets loads cached application assets.
package assets

import (
	"context"
	"errors"
	"io"
	stdhttp "net/http"
	"sync"

	"github.com/rfwlab/rfw/v2/http"
	"github.com/rfwlab/rfw/v2/internal/safehttp"
)

var httpClient = safehttp.NewClient()
var httpClientMu sync.RWMutex

// SetNativeClient replaces the native asset client. Custom clients may reach
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

func fetch(url string) (*stdhttp.Response, error) {
	req, err := safehttp.NewRequest(context.Background(), stdhttp.MethodGet, url)
	if err != nil {
		return nil, err
	}
	return currentNativeClient().Do(req)
}

// Image is a placeholder for non-WASM builds.
type Image struct {
	URL  string
	Data []byte
}

var loadImageFn = func(url string, done func(Image, error)) {
	resp, err := fetch(url)
	if err != nil {
		done(Image{}, err)
		return
	}

	b, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil {
		done(Image{}, err)
		return
	}
	if closeErr != nil {
		done(Image{}, closeErr)
		return
	}
	if resp.StatusCode >= 400 {
		done(Image{}, errors.New(string(b)))
		return
	}
	done(Image{URL: url, Data: b}, nil)
}

var loadBinaryFn = func(url string, done func([]byte, error)) {
	resp, err := fetch(url)
	if err != nil {
		done(nil, err)
		return
	}

	b, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil {
		done(nil, err)
		return
	}
	if closeErr != nil {
		done(nil, closeErr)
		return
	}
	if resp.StatusCode >= 400 {
		done(nil, errors.New(string(b)))
		return
	}
	done(b, nil)
}

type imageEntry struct {
	once  sync.Once
	img   Image
	err   error
	ready chan struct{}
}

var imageCache sync.Map // map[string]*imageEntry

// LoadImage starts or reads a cached image request.
func LoadImage(url string) (Image, error) {
	ceIface, _ := imageCache.LoadOrStore(url, &imageEntry{ready: make(chan struct{})})
	ce := ceIface.(*imageEntry)

	ce.once.Do(func() {
		go loadImageFn(url, func(v Image, err error) {
			ce.img = v
			ce.err = err
			close(ce.ready)
		})
	})

	select {
	case <-ce.ready:
		if ce.err != nil {
			return Image{}, ce.err
		}
		return ce.img, nil
	default:
		return Image{}, http.ErrPending
	}
}

type modelEntry struct {
	once  sync.Once
	data  []byte
	err   error
	ready chan struct{}
}

var modelCache sync.Map // map[string]*modelEntry

// LoadModel starts or reads a cached binary model request.
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

// LoadJSON retrieves and decodes JSON from url.
func LoadJSON(url string, v any) error { return http.FetchJSON(url, v) }

// ClearCache removes cached assets for url.
func ClearCache(url string) {
	imageCache.Delete(url)
	modelCache.Delete(url)
}
