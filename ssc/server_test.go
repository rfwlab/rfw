//go:build !js

package ssc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/host"
	"golang.org/x/net/websocket"
)

func TestSSCEventBus(t *testing.T) {
	type seenEvent struct {
		component string
		value     any
	}
	seen := make(chan seenEvent, 1)
	SubscribeSSC(func(_ context.Context, e Event) error {
		seen <- seenEvent{component: e.Component, value: e.Payload["value"]}
		return nil
	})

	if err := EmitSSC(context.Background(), Event{Component: "Counter", Payload: map[string]any{"value": 2}}); err != nil {
		t.Fatalf("emit failed: %v", err)
	}

	got := <-seen
	if got.component != "Counter" || got.value != 2 {
		t.Fatalf("unexpected event: %+v", got)
	}
}

func TestSSCServerServesIndexAndWasmHeaders(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<main>app</main>"), 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.wasm.br"), []byte("wasm"), 0o600); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "rfw_config.js"), []byte("//cfg"), 0o600); err != nil {
		t.Fatalf("write runtime config: %v", err)
	}

	server := NewSSCServer(":0", root)
	ts := httptest.NewServer(server.Mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/docs/anything")
	if err != nil {
		t.Fatalf("index fallback request failed: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close index response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected index fallback 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/app.wasm.br?v=abc123")
	if err != nil {
		t.Fatalf("wasm request failed: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close versioned wasm response: %v", err)
	}
	if resp.Header.Get("Content-Encoding") != "br" {
		t.Fatalf("expected br encoding, got %q", resp.Header.Get("Content-Encoding"))
	}
	if resp.Header.Get("Content-Type") != "application/wasm" {
		t.Fatalf("expected wasm content type, got %q", resp.Header.Get("Content-Type"))
	}
	if cache := resp.Header.Get("Cache-Control"); cache != "public, max-age=31536000, immutable" {
		t.Fatalf("unexpected Cache-Control header: %q", cache)
	}

	resp, err = http.Get(ts.URL + "/app.wasm.br?v=")
	if err != nil {
		t.Fatalf("unversioned wasm request failed: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close unversioned wasm response: %v", err)
	}
	if cache := resp.Header.Get("Cache-Control"); cache != "no-cache" {
		t.Fatalf("unexpected unversioned Cache-Control header: %q", cache)
	}

	resp, err = http.Get(ts.URL + "/rfw_config.js")
	if err != nil {
		t.Fatalf("runtime config request failed: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close runtime config response: %v", err)
	}
	if cache := resp.Header.Get("Cache-Control"); cache != "no-cache" {
		t.Fatalf("unexpected runtime config Cache-Control header: %q", cache)
	}
}

func TestSSCWithSessionTargetDelegatesHostOption(t *testing.T) {
	var opts host.BroadcastOptions
	WithSessionTarget("abc")(&opts)
	if opts.Session != "abc" {
		t.Fatalf("expected session abc, got %q", opts.Session)
	}
}

func TestSSCServerDevModeDoesNotCacheAssets(t *testing.T) {
	t.Setenv("RFW_DEV_BUILD", "1")
	root := t.TempDir()
	for name, contents := range map[string]string{
		"index.html":    "<main>app</main>",
		"app.wasm":      "wasm",
		"rfw_config.js": "//cfg",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	ts := httptest.NewServer(NewSSCServer(":0", root).Mux)
	defer ts.Close()
	for _, path := range []string{"/", "/app.wasm?v=abc123", "/rfw_config.js"} {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatalf("get %s: %v", path, err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("close %s: %v", path, err)
		}
		if cache := resp.Header.Get("Cache-Control"); cache != "no-store" {
			t.Fatalf("%s Cache-Control = %q, want no-store", path, cache)
		}
	}
}

// The /ws endpoint honours host.MuxOption guards; by default it stays open.
func TestSSCServerWSOriginAllowlist(t *testing.T) {
	root := t.TempDir()
	s := NewSSCServer(":0", root, host.WithOriginAllowlist("https://app.example.com"))
	srv := httptest.NewServer(s.Mux)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/ws", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.Header.Set("Origin", "http://evil.example")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close origin response: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for unlisted origin, got %d", resp.StatusCode)
	}
}

func TestSSCServerResumesAtSessionLimit(t *testing.T) {
	type request struct{}
	type response struct {
		Count int `json:"count"`
	}
	const action = "test.ssc.resume.limit"
	if err := host.RegisterAction(action, func(_ context.Context, session *host.Session, _ request) (response, error) {
		count := 0
		if stored, ok := session.ContextGet("count"); ok {
			count = stored.(int)
		}
		count++
		session.ContextSet("count", count)
		return response{Count: count}, nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}

	server := httptest.NewServer(NewSSCServer(":0", t.TempDir(), host.WithSSCLimits(host.SSCLimits{
		MaxSessions:    1,
		ResumeTTL:      time.Second,
		ReplayMessages: 8,
	})).Mux)
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	dial := func() *websocket.Conn {
		socket, err := websocket.Dial(wsURL, "", server.URL)
		if err != nil {
			t.Fatalf("dial websocket: %v", err)
		}
		return socket
	}
	send := func(socket *websocket.Conn, message host.Inbound) {
		data, err := json.Marshal(message)
		if err != nil {
			t.Fatalf("marshal message: %v", err)
		}
		if err := websocket.Message.Send(socket, data); err != nil {
			t.Fatalf("send message: %v", err)
		}
	}
	receive := func(socket *websocket.Conn) host.Outbound {
		var data []byte
		if err := websocket.Message.Receive(socket, &data); err != nil {
			t.Fatalf("receive message: %v", err)
		}
		var message host.Outbound
		if err := json.Unmarshal(data, &message); err != nil {
			t.Fatalf("decode message: %v", err)
		}
		return message
	}

	firstSocket := dial()
	send(firstSocket, host.Inbound{Action: action, ID: "first", Sequence: 1})
	first := receive(firstSocket)
	closeTestResource(t, firstSocket)

	var (
		secondSocket *websocket.Conn
		second       host.Outbound
	)
	deadline := time.Now().Add(time.Second)
	for {
		secondSocket = dial()
		send(secondSocket, host.Inbound{
			Action:      action,
			ID:          "second",
			Sequence:    2,
			Ack:         first.Sequence,
			ResumeToken: first.ResumeToken,
		})
		second = receive(secondSocket)
		if second.Session == first.Session && second.ID == "second" {
			break
		}
		closeTestResource(t, secondSocket)
		if time.Now().After(deadline) {
			t.Fatalf("session did not resume at the limit: first=%#v second=%#v", first, second)
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer closeTestResource(t, secondSocket)
	payload, ok := second.Payload.(map[string]any)
	if !ok || payload["count"] != float64(2) {
		t.Fatalf("session state was not retained: %#v", second.Payload)
	}
	if session, ok := host.SessionByID(first.Session); ok {
		host.ReleaseSession(session)
	}
}
