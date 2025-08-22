package host

import (
	"encoding/json"
	"io"
	"log"

	"golang.org/x/net/websocket"
)

type inbound struct {
	Component string         `json:"component"`
	Payload   map[string]any `json:"payload"`
}

func wsHandler(ws *websocket.Conn) {
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
		if hc, ok := Get(msg.Component); ok {
			if resp := hc.Handle(msg.Payload); resp != nil {
				out := struct {
					Component string `json:"component"`
					Payload   any    `json:"payload"`
				}{Component: msg.Component, Payload: resp}
				b, err := json.Marshal(out)
				if err == nil {
					if err := websocket.Message.Send(ws, b); err != nil {
						log.Printf("send: %v", err)
					}
				}
			}
		}
	}
}
