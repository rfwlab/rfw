//go:build !js || !wasm

package assets

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestFetchRejectsNonHTTPURLs(t *testing.T) {
	for _, rawURL := range []string{"file:///tmp/secret", "/relative", "http://127.0.0.1/secret"} {
		resp, err := fetch(rawURL)
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Fatalf("close response for %q: %v", rawURL, closeErr)
			}
		}
		if err == nil {
			t.Fatalf("expected %q to be rejected", rawURL)
		}
		if strings.HasPrefix(rawURL, "http://127.") && !strings.Contains(err.Error(), "not public") {
			t.Fatalf("unexpected private URL rejection: %v", err)
		}
	}
}

func TestLoadModel_CacheAndPending(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(200)
		_, _ = w.Write([]byte{1, 2, 3})
	}))
	defer srv.Close()
	oldClient := currentNativeClient()
	SetNativeClient(srv.Client())
	t.Cleanup(func() { SetNativeClient(oldClient) })

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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		time.Sleep(15 * time.Millisecond)
		w.WriteHeader(200)
		_, _ = w.Write([]byte("PNGDATA"))
	}))
	defer srv.Close()
	oldClient := currentNativeClient()
	SetNativeClient(srv.Client())
	t.Cleanup(func() { SetNativeClient(oldClient) })

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

func TestLoadJSONRejectsPrivateNetwork(t *testing.T) {
	const privateURL = "http://127.0.0.1/data.json"
	v1http.ClearCache(privateURL)
	t.Cleanup(func() { v1http.ClearCache(privateURL) })
	var out struct {
		V int `json:"v"`
	}
	if err := LoadJSON(privateURL, &out); err != v1http.ErrPending {
		t.Fatalf("expected ErrPending, got %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		err := LoadJSON(privateURL, &out)
		if err != nil && err != v1http.ErrPending {
			if !strings.Contains(err.Error(), "not public") {
				t.Fatalf("unexpected private URL rejection: %v", err)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for private URL rejection")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
