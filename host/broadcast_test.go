package host

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestClientCanUnsubscribeFromBroadcasts(t *testing.T) {
	const componentName = "broadcast-unsubscribe"
	Register(NewHostComponent(componentName, func(map[string]any) any {
		return map[string]any{"ready": true}
	}))

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("ok"), 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}
	srv := httptest.NewServer(NewMux(root))
	defer srv.Close()
	ws, err := websocket.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", "", srv.URL)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer closeTestResource(t, ws)

	init, _ := json.Marshal(map[string]any{"component": componentName, "payload": map[string]any{"init": true}})
	if err := websocket.Message.Send(ws, init); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	var raw []byte
	if err := websocket.Message.Receive(ws, &raw); err != nil {
		t.Fatalf("receive init: %v", err)
	}

	unsubscribe, _ := json.Marshal(map[string]any{"component": componentName, "payload": map[string]any{"unsubscribe": true}})
	if err := websocket.Message.Send(ws, unsubscribe); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}
	if err := websocket.Message.Receive(ws, &raw); err != nil {
		t.Fatalf("receive unsubscribe acknowledgement: %v", err)
	}

	Broadcast(componentName, map[string]any{"unexpected": true})
	if err := ws.SetDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	if err := websocket.Message.Receive(ws, &raw); err == nil {
		t.Fatalf("received broadcast after unsubscribe: %s", raw)
	}
	if err := ws.SetDeadline(time.Time{}); err != nil {
		t.Fatalf("reset deadline: %v", err)
	}
}

// Broadcast must snapshot the (conn, session) pairs under connMu: clients
// subscribing and disconnecting concurrently with broadcasts used to race with
// the map iteration. Run with -race to exercise the invariant.
func TestBroadcastConcurrentWithConnectionChurn(t *testing.T) {
	const componentName = "broadcast-churn"
	Register(NewHostComponentWithSession(componentName, func(_ *Session, _ map[string]any) any {
		return map[string]any{"ok": true}
	}))

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("ok"), 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}
	srv := httptest.NewServer(NewMux(root))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	init, err := json.Marshal(map[string]any{
		"component": componentName,
		"payload":   map[string]any{"init": true},
	})
	if err != nil {
		t.Fatalf("marshal init: %v", err)
	}

	var wg sync.WaitGroup
	// Churn: connections subscribe, receive the handler response and close.
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				ws, err := websocket.Dial(wsURL, "", srv.URL)
				if err != nil {
					continue
				}
				if err := websocket.Message.Send(ws, init); err == nil {
					var raw []byte
					_ = websocket.Message.Receive(ws, &raw)
				}
				closeTestResource(t, ws)
			}
		}()
	}
	// Broadcasters run against the same component while the map churns.
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				Broadcast(componentName, map[string]any{"tick": j})
			}
		}()
	}
	wg.Wait()
}
