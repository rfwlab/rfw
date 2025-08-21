package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"

	"github.com/rfwlab/rfw/v1/host"
)

type inbound struct {
	Component string          `json:"component"`
	Payload   json.RawMessage `json:"payload"`
}

func main() {
	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			var raw []byte
			if err := websocket.Message.Receive(ws, &raw); err != nil {
				if err == io.EOF {
					break
				}
				log.Printf("recv: %v", err)
				return
			}
			var msg inbound
			if err := json.Unmarshal(raw, &msg); err != nil {
				log.Printf("unmarshal: %v", err)
				continue
			}
			if hc, ok := host.Get(msg.Component); ok {
				ctx := host.NewContext(ws, msg.Component)
				hc.Handle(ctx, msg.Payload)
			}
		}
	}))
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func init() {
	host.Register(host.NewHostComponent("HomeHost", func(ctx *host.Context, payload json.RawMessage) {
		_ = ctx.Notify(map[string]any{"welcome": "hello from host"})
	}))
}
