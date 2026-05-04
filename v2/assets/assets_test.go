package assets

import (
"net/http"
"net/http/httptest"
"testing"
"time"

v1http "github.com/rfwlab/rfw/v2/http"
)

func waitImage(t *testing.T, fn func() (Image, error)) Image {
t.Helper()
deadline := time.Now().Add(2 * time.Second)
for {
img, err := fn()
if err == nil {
return img
}
if err != v1http.ErrPending {
t.Fatalf("unexpected error: %v", err)
}
if time.Now().After(deadline) {
t.Fatalf("timed out waiting for image")
}
time.Sleep(5 * time.Millisecond)
}
}

func waitBytes(t *testing.T, fn func() ([]byte, error)) []byte {
t.Helper()
deadline := time.Now().Add(2 * time.Second)
for {
b, err := fn()
if err == nil {
return b
}
if err != v1http.ErrPending {
t.Fatalf("unexpected error: %v", err)
}
if time.Now().After(deadline) {
t.Fatalf("timed out waiting for bytes")
}
time.Sleep(5 * time.Millisecond)
}
}

func TestLoadModel_CacheAndPending(t *testing.T) {
var hits int
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
hits++
time.Sleep(20 * time.Millisecond)
w.WriteHeader(200)
_, _ = w.Write([]byte{1, 2, 3})
}))
defer srv.Close()

ClearCache(srv.URL)
t.Cleanup(func() { ClearCache(srv.URL) })

if _, err := LoadModel(srv.URL); err != v1http.ErrPending {
t.Fatalf("expected ErrPending, got %v", err)
}

got := waitBytes(t, func() ([]byte, error) { return LoadModel(srv.URL) })
if len(got) != 3 || got[0] != 1 || got[2] != 3 {
t.Fatalf("unexpected bytes: %v", got)
}

got2, err := LoadModel(srv.URL)
if err != nil {
t.Fatalf("expected cached success, got %v", err)
}
if len(got2) != 3 || got2[1] != 2 {
t.Fatalf("unexpected cached bytes: %v", got2)
}

if hits != 1 {
t.Fatalf("expected 1 server hit, got %d", hits)
}
}

func TestLoadImage_UsesCache(t *testing.T) {
var hits int
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
hits++
time.Sleep(15 * time.Millisecond)
w.WriteHeader(200)
_, _ = w.Write([]byte("PNGDATA"))
}))
defer srv.Close()

ClearCache(srv.URL)
t.Cleanup(func() { ClearCache(srv.URL) })

if _, err := LoadImage(srv.URL); err != v1http.ErrPending {
t.Fatalf("expected ErrPending, got %v", err)
}

img := waitImage(t, func() (Image, error) { return LoadImage(srv.URL) })
if img.URL != srv.URL || string(img.Data) != "PNGDATA" {
t.Fatalf("unexpected image: %+v", img)
}

img2, err := LoadImage(srv.URL)
if err != nil {
t.Fatalf("expected cached image, got %v", err)
}
if string(img2.Data) != "PNGDATA" {
t.Fatalf("unexpected cached data: %q", string(img2.Data))
}

if hits != 1 {
t.Fatalf("expected 1 hit, got %d", hits)
}
}

func TestLoadJSON_DelegatesToHTTP(t *testing.T) {
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
time.Sleep(10 * time.Millisecond)
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(200)
_, _ = w.Write([]byte(`{"v": 7}`))
}))
defer srv.Close()

var out struct{ V int `json:"v"` }
if err := LoadJSON(srv.URL, &out); err != v1http.ErrPending {
t.Fatalf("expected ErrPending, got %v", err)
}

deadline := time.Now().Add(2 * time.Second)
for {
err := LoadJSON(srv.URL, &out)
if err == nil {
break
}
if err != v1http.ErrPending {
t.Fatalf("unexpected error: %v", err)
}
if time.Now().After(deadline) {
t.Fatalf("timed out waiting for json")
}
time.Sleep(5 * time.Millisecond)
}

if out.V != 7 {
t.Fatalf("expected V=7, got %d", out.V)
}
}
