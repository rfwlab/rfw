//go:build !js || !wasm

package assets

import (
"errors"
"io"
stdhttp "net/http"
"sync"

"github.com/rfwlab/rfw/v1/http"
)

// Image is a placeholder for non-WASM builds.
type Image struct {
URL  string
Data []byte
}

var loadImageFn = func(url string, done func(Image, error)) {
resp, err := stdhttp.Get(url)
if err != nil {
done(Image{}, err)
return
}
defer resp.Body.Close()

b, err := io.ReadAll(resp.Body)
if err != nil {
done(Image{}, err)
return
}
if resp.StatusCode >= 400 {
done(Image{}, errors.New(string(b)))
return
}
done(Image{URL: url, Data: b}, nil)
}

var loadBinaryFn = func(url string, done func([]byte, error)) {
resp, err := stdhttp.Get(url)
if err != nil {
done(nil, err)
return
}
defer resp.Body.Close()

b, err := io.ReadAll(resp.Body)
if err != nil {
done(nil, err)
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

func LoadJSON(url string, v any) error { return http.FetchJSON(url, v) }

func ClearCache(url string) {
imageCache.Delete(url)
modelCache.Delete(url)
}
