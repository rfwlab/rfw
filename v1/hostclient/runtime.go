//go:build js && wasm

package hostclient

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	dom "github.com/rfwlab/rfw/v1/dom"
	js "github.com/rfwlab/rfw/v1/js"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type componentBinding struct {
	id   string
	vars []string
}

var (
	conn     *websocket.Conn
	bindings = map[string]componentBinding{}
	once     sync.Once
	mu       sync.RWMutex
	pending  []message
	handlers = map[string]func(map[string]any){}
	debug    bool

	sessionMu sync.RWMutex
	sessionID string
)

type message struct {
	name    string
	payload any
}

func connect() {
	once.Do(func() {
		go connectionLoop()
	})
}

func connectionLoop() {
	host := js.Location().Get("host").String()
	if h := js.Get("RFW_HOST"); h.Truthy() {
		host = h.String()
	}
	scheme := "wss"
	if js.Location().Get("protocol").String() == "http:" {
		scheme = "ws"
	}
	backoff := time.Second
	for {
		ctx, cancel := context.WithCancel(context.Background())
		url := fmt.Sprintf("%s://%s/ws", scheme, host)
		if debug {
			log.Printf("hostclient: dialing %s", url)
		}
		c, _, err := websocket.Dial(ctx, url, nil)
		if err != nil {
			cancel()
			time.Sleep(backoff)
			if backoff < time.Minute {
				backoff *= 2
			}
			continue
		}

		mu.Lock()
		conn = c
		pend := pending
		pending = nil
		mu.Unlock()

		backoff = time.Second
		if debug {
			log.Printf("hostclient: connected")
		}

		for _, msg := range pend {
			sendMessage(c, msg)
		}
		for name := range bindings {
			sendMessage(c, message{name: name, payload: map[string]any{"init": true}})
		}

		errCh := make(chan error, 1)
		go func() { errCh <- readLoop(ctx, c) }()
		go func() { errCh <- pingLoop(ctx, c) }()
		err = <-errCh
		cancel()
		c.Close(websocket.StatusInternalError, "connection closed")

		mu.Lock()
		conn = nil
		mu.Unlock()
		if debug && err != nil {
			log.Printf("hostclient: connection closed: %v", err)
		}
	}
}

func pingLoop(ctx context.Context, c *websocket.Conn) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Ping(pctx)
			cancel()
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func readLoop(ctx context.Context, c *websocket.Conn) error {
	for {
		var msg struct {
			Component string         `json:"component"`
			Payload   map[string]any `json:"payload"`
			Session   string         `json:"session"`
		}
		if err := wsjson.Read(ctx, c, &msg); err != nil {
			return err
		}
		if debug {
			log.Printf("hostclient: recv %s %v", msg.Component, msg.Payload)
		}
		if msg.Session != "" {
			sessionMu.Lock()
			sessionID = msg.Session
			sessionMu.Unlock()
		}
		if h, ok := handlers[msg.Component]; ok {
			payload := msg.Payload
			if payload == nil {
				payload = make(map[string]any)
			}
			if msg.Session != "" {
				payload["_session"] = msg.Session
			}
			h(payload)
			continue
		}
		if b, ok := bindings[msg.Component]; ok {
			rootEl := dom.Doc().Query(fmt.Sprintf("[data-component-id='%s']", b.id))
			if !rootEl.Truthy() {
				continue
			}
			root := newComponentRoot(rootEl)

			if snap := decodeInitSnapshotPayload(msg.Payload["initSnapshot"]); snap != nil {
				applyInitSnapshot(root, snap)
				if len(snap.Vars) > 0 {
					b.vars = append([]string(nil), snap.Vars...)
					bindings[msg.Component] = b
				}
				continue
			}

			mismatches := handleHostPayload(root, msg.Payload)
			if len(mismatches) > 0 {
				for _, m := range mismatches {
					log.Printf("hostclient: hydration mismatch component=%s var=%s expected=%s actualHash=%s actual=%q", msg.Component, m.VarName, m.Expected, m.ActualHash, m.Actual)
				}
				Send(msg.Component, buildResyncPayload(mismatches))
			}
		}
	}
}

func RegisterComponent(id, name string, vars []string) {
	bindings[name] = componentBinding{id: id, vars: vars}
	connect()
	Send(name, map[string]any{"init": true})
}

func Send(name string, payload any) {
	connect()
	mu.RLock()
	c := conn
	mu.RUnlock()
	if c == nil {
		mu.Lock()
		pending = append(pending, message{name: name, payload: payload})
		mu.Unlock()
		return
	}
	if debug {
		log.Printf("hostclient: send %s %v", name, payload)
	}
	sendMessage(c, message{name: name, payload: payload})
}

// RegisterHandler installs a callback for messages targeting the component name.
func RegisterHandler(name string, h func(map[string]any)) {
	mu.Lock()
	handlers[name] = h
	mu.Unlock()
	connect()
	Send(name, map[string]any{"init": true})
}

// SessionID returns the current WebSocket session identifier assigned by the host.
func SessionID() string {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return sessionID
}

func sendMessage(c *websocket.Conn, msg message) {
	m := struct {
		Component string `json:"component"`
		Payload   any    `json:"payload"`
	}{Component: msg.name, Payload: msg.payload}
	ctx := context.Background()
	_ = wsjson.Write(ctx, c, m)
}

// EnableDebug prints WebSocket traffic to the console.
func EnableDebug() { debug = true }
