package host

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	"golang.org/x/net/websocket"
)

type inbound struct {
	Component string         `json:"component"`
	Payload   map[string]any `json:"payload"`
}

var (
	connections = make(map[string]map[*websocket.Conn]struct{})
	connMu      sync.Mutex
)

func wsHandler(ws *websocket.Conn) {
	var subscribed []string
	defer func() {
		connMu.Lock()
		for _, name := range subscribed {
			if set, ok := connections[name]; ok {
				delete(set, ws)
				if len(set) == 0 {
					delete(connections, name)
				}
			}
		}
		connMu.Unlock()
		ws.Close()
	}()
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
			connMu.Lock()
			if _, ok := connections[msg.Component]; !ok {
				connections[msg.Component] = make(map[*websocket.Conn]struct{})
			}
			connections[msg.Component][ws] = struct{}{}
			subscribed = append(subscribed, msg.Component)
			connMu.Unlock()
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

// Broadcast sends the given payload to all connections subscribed to the component name.
func Broadcast(name string, payload any) {
	connMu.Lock()
	conns := connections[name]
	connMu.Unlock()
	if len(conns) == 0 {
		return
	}
	out := struct {
		Component string `json:"component"`
		Payload   any    `json:"payload"`
	}{Component: name, Payload: payload}
	b, err := json.Marshal(out)
	if err != nil {
		return
	}
	for ws := range conns {
		if err := websocket.Message.Send(ws, b); err != nil {
			log.Printf("send: %v", err)
		}
	}
}
