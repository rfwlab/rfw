//go:build js && wasm

package hostclient

import (
	"context"
	"fmt"
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
	host := js.Global().Get("location").Get("host").String()
	if h := js.Global().Get("RFW_HOST"); h.Truthy() {
		host = h.String()
	}
	backoff := time.Second
	for {
		ctx, cancel := context.WithCancel(context.Background())
		c, _, err := websocket.Dial(ctx, fmt.Sprintf("wss://%s/ws", host), nil)
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

		for _, msg := range pend {
			sendMessage(c, msg)
		}
		for name := range bindings {
			sendMessage(c, message{name: name, payload: map[string]any{"init": true}})
		}

		errCh := make(chan error, 1)
		go func() { errCh <- readLoop(ctx, c) }()
		go func() { errCh <- pingLoop(ctx, c) }()
		<-errCh
		cancel()
		c.Close(websocket.StatusInternalError, "connection closed")

		mu.Lock()
		conn = nil
		mu.Unlock()
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
		}
		if err := wsjson.Read(ctx, c, &msg); err != nil {
			return err
		}
		if b, ok := bindings[msg.Component]; ok {
			root := dom.Query(fmt.Sprintf("[data-component-id='%s']", b.id))
			if root.Truthy() {
				for k, v := range msg.Payload {
					el := root.Call("querySelector", fmt.Sprintf(`[data-host-var="%s"]`, k))
					if el.Truthy() {
						el.Set("textContent", fmt.Sprintf("%v", v))
					}
				}
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
	sendMessage(c, message{name: name, payload: payload})
}

func sendMessage(c *websocket.Conn, msg message) {
	m := struct {
		Component string `json:"component"`
		Payload   any    `json:"payload"`
	}{Component: msg.name, Payload: msg.payload}
	ctx := context.Background()
	_ = wsjson.Write(ctx, c, m)
}
