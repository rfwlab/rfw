package http

import (
"net/http"
"net/http/httptest"
"sync"
"testing"
"time"
)

func waitNoPending(t *testing.T, fn func() error) {
t.Helper()
deadline := time.Now().Add(2 * time.Second)
for {
err := fn()
if err == nil {
return
}
if err != ErrPending {
t.Fatalf("unexpected error: %v", err)
}
if time.Now().After(deadline) {
t.Fatalf("timed out waiting for request to complete")
}
time.Sleep(5 * time.Millisecond)
}
}

func waitText(t *testing.T, fn func() (string, error)) string {
t.Helper()
deadline := time.Now().Add(2 * time.Second)
for {
s, err := fn()
if err == nil {
return s
}
if err != ErrPending {
t.Fatalf("unexpected error: %v", err)
}
if time.Now().After(deadline) {
t.Fatalf("timed out waiting for request to complete")
}
time.Sleep(5 * time.Millisecond)
}
}

func TestFetchText_CacheAndPending(t *testing.T) {
var hits int
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
hits++
time.Sleep(30 * time.Millisecond)
w.WriteHeader(200)
_, _ = w.Write([]byte("hello"))
}))
defer srv.Close()

ClearCache(srv.URL)
t.Cleanup(func() { ClearCache(srv.URL) })

if _, err := FetchText(srv.URL); err != ErrPending {
t.Fatalf("expected ErrPending, got %v", err)
}

got := waitText(t, func() (string, error) { return FetchText(srv.URL) })
if got != "hello" {
t.Fatalf("expected 'hello', got %q", got)
}

got2, err := FetchText(srv.URL)
if err != nil {
t.Fatalf("expected cached success, got %v", err)
}
if got2 != "hello" {
t.Fatalf("expected cached 'hello', got %q", got2)
}
if hits != 1 {
t.Fatalf("expected 1 hit, got %d", hits)
}
}

func TestFetchJSON_DecodeAndHook(t *testing.T) {
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
time.Sleep(20 * time.Millisecond)
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(200)
_, _ = w.Write([]byte(`{"ok":true,"n":3}`))
}))
defer srv.Close()

ClearCache(srv.URL)
t.Cleanup(func() { ClearCache(srv.URL) })

var mu sync.Mutex
var starts, completes int
var gotStatus int
RegisterHTTPHook(func(start bool, _ string, status int, d time.Duration) {
mu.Lock()
defer mu.Unlock()
if start {
starts++
return
}
completes++
gotStatus = status
if d <= 0 {
t.Fatalf("expected duration > 0")
}
})
t.Cleanup(func() { RegisterHTTPHook(nil) })

var out struct {
OK bool `json:"ok"`
N  int  `json:"n"`
}

if err := FetchJSON(srv.URL, &out); err != ErrPending {
t.Fatalf("expected ErrPending, got %v", err)
}

waitNoPending(t, func() error { return FetchJSON(srv.URL, &out) })
if !out.OK || out.N != 3 {
t.Fatalf("unexpected decoded value: %+v", out)
}

mu.Lock()
defer mu.Unlock()
if starts != 1 || completes != 1 {
t.Fatalf("expected 1 start and 1 complete, got %d and %d", starts, completes)
}
if gotStatus != 200 {
t.Fatalf("expected status 200, got %d", gotStatus)
}
}

func TestClearCache_AllowsRefetch(t *testing.T) {
var hits int
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
hits++
w.WriteHeader(200)
_, _ = w.Write([]byte("ok"))
}))
defer srv.Close()

ClearCache(srv.URL)
_ = waitText(t, func() (string, error) { return FetchText(srv.URL) })

ClearCache(srv.URL)
_ = waitText(t, func() (string, error) { return FetchText(srv.URL) })

if hits != 2 {
t.Fatalf("expected 2 hits after clear, got %d", hits)
}
}
