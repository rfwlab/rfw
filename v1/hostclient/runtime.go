//go:build js && wasm

package hostclient

import (
	"context"
	"fmt"
	"sync"

	dom "github.com/rfwlab/rfw/v1/dom"
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
)

func connect() {
	once.Do(func() {
		ctx := context.Background()
		c, _, err := websocket.Dial(ctx, "ws://localhost:8090/ws", nil)
		if err != nil {
			return
		}
		conn = c
		go readLoop(ctx)
	})
}

func readLoop(ctx context.Context) {
	for {
		var msg struct {
			Component string         `json:"component"`
			Payload   map[string]any `json:"payload"`
		}
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return
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
	msg := struct {
		Component string `json:"component"`
		Payload   any    `json:"payload"`
	}{Component: name, Payload: payload}
	ctx := context.Background()
	_ = wsjson.Write(ctx, conn, msg)
}
