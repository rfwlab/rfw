package host

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"golang.org/x/net/websocket"

	"github.com/rfwlab/rfw/v1/state"
)

func TestSessionIsolation(t *testing.T) {
	t.Helper()

	registry = make(map[string]*HostComponent)

	const componentName = "SessionHost"
	Register(NewHostComponentWithSession(componentName, func(session *Session, payload map[string]any) any {
		const storeKey = "counter"
		storeVal, ok := session.ContextGet(storeKey)
		var store *state.Store
		if ok {
			store = storeVal.(*state.Store)
		} else {
			store = session.StoreManager().NewStore("counter")
			store.Set("value", 0)
			session.ContextSet(storeKey, store)
		}
		if inc, ok := payload["increment"].(bool); ok && inc {
			current, _ := store.Get("value").(int)
			store.Set("value", current+1)
		}
		return map[string]any{"value": store.Get("value")}
	}))

	root := t.TempDir()
	// Ensure an index exists so NewMux can serve fallback responses without error.
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	srv := httptest.NewServer(loggingMiddleware(NewMux(root)))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	type sessionConn struct {
		ws  *websocket.Conn
		id  string
		idx int
	}

	dial := func(idx int) sessionConn {
		ws, err := websocket.Dial(wsURL, "", srv.URL)
		if err != nil {
			t.Fatalf("dial %d: %v", idx, err)
		}
		init := map[string]any{
			"component": componentName,
			"payload":   map[string]any{"init": true},
		}
		raw, err := json.Marshal(init)
		if err != nil {
			t.Fatalf("marshal init %d: %v", idx, err)
		}
		if err := websocket.Message.Send(ws, raw); err != nil {
			t.Fatalf("send init %d: %v", idx, err)
		}
		var respRaw []byte
		if err := websocket.Message.Receive(ws, &respRaw); err != nil {
			t.Fatalf("recv init %d: %v", idx, err)
		}
		var resp struct {
			Component string         `json:"component"`
			Payload   map[string]any `json:"payload"`
			Session   string         `json:"session"`
		}
		if err := json.Unmarshal(respRaw, &resp); err != nil {
			t.Fatalf("unmarshal init %d: %v", idx, err)
		}
		if resp.Session == "" {
			t.Fatalf("session id missing for conn %d", idx)
		}
		if resp.Component != componentName {
			t.Fatalf("unexpected component %s", resp.Component)
		}
		if val, ok := resp.Payload["value"].(float64); !ok || val != 0 {
			t.Fatalf("unexpected init value for conn %d: %v", idx, resp.Payload)
		}
		return sessionConn{ws: ws, id: resp.Session, idx: idx}
	}

	sessions := []sessionConn{dial(0), dial(1)}
	defer func() {
		for _, sc := range sessions {
			sc.ws.Close()
		}
	}()

	counts := []int{5, 2}
	if len(counts) != len(sessions) {
		t.Fatalf("mismatched counts")
	}

	errCh := make(chan error, len(sessions))
	var wg sync.WaitGroup
	for i, sc := range sessions {
		wg.Add(1)
		count := counts[i]
		go func(sc sessionConn, target int) {
			defer wg.Done()
			for j := 0; j < target; j++ {
				payload := map[string]any{
					"component": componentName,
					"payload":   map[string]any{"increment": true},
				}
				raw, err := json.Marshal(payload)
				if err != nil {
					errCh <- fmt.Errorf("marshal increment idx=%d: %w", sc.idx, err)
					return
				}
				if err := websocket.Message.Send(sc.ws, raw); err != nil {
					errCh <- fmt.Errorf("send increment idx=%d: %w", sc.idx, err)
					return
				}
				var respRaw []byte
				if err := websocket.Message.Receive(sc.ws, &respRaw); err != nil {
					errCh <- fmt.Errorf("recv increment idx=%d: %w", sc.idx, err)
					return
				}
				var resp struct {
					Component string         `json:"component"`
					Payload   map[string]any `json:"payload"`
					Session   string         `json:"session"`
				}
				if err := json.Unmarshal(respRaw, &resp); err != nil {
					errCh <- fmt.Errorf("unmarshal increment idx=%d: %w", sc.idx, err)
					return
				}
				if resp.Session != sc.id {
					errCh <- fmt.Errorf("response session mismatch idx=%d: got %s want %s", sc.idx, resp.Session, sc.id)
					return
				}
			}
			errCh <- nil
		}(sc, count)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}

	for i, sc := range sessions {
		sess, ok := SessionByID(sc.id)
		if !ok {
			t.Fatalf("session %d not found", i)
		}
		snap := sess.Snapshot()
		module := snap["default"]
		if module == nil {
			t.Fatalf("session %d missing default module snapshot", i)
		}
		counter := module["counter"]
		if counter == nil {
			t.Fatalf("session %d missing counter store", i)
		}
		val, ok := counter["value"].(int)
		if !ok {
			t.Fatalf("session %d missing value entry: %v", i, counter)
		}
		if val != counts[i] {
			t.Fatalf("session %d got value %d want %d", i, val, counts[i])
		}
	}

	if sessions[0].id == sessions[1].id {
		t.Fatal("session ids should differ")
	}
}
