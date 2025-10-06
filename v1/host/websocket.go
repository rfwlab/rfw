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

type outbound struct {
	Component string `json:"component"`
	Payload   any    `json:"payload,omitempty"`
	Session   string `json:"session"`
}

type broadcastOptions struct {
	session string
}

// BroadcastOption configures a broadcast call.
type BroadcastOption func(*broadcastOptions)

// WithSessionTarget limits a broadcast to a specific session ID.
func WithSessionTarget(sessionID string) BroadcastOption {
	return func(opts *broadcastOptions) {
		opts.session = sessionID
	}
}

var (
	connections = make(map[string]map[*websocket.Conn]*Session)
	connMu      sync.RWMutex
)

func wsHandler(ws *websocket.Conn) {
	session := allocateSession()
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
		releaseSession(session)
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
				connections[msg.Component] = make(map[*websocket.Conn]*Session)
			}
			if _, tracked := connections[msg.Component][ws]; !tracked {
				connections[msg.Component][ws] = session
				subscribed = append(subscribed, msg.Component)
			}
			connMu.Unlock()
			resp := hc.HandleWithSession(session, msg.Payload)
			if resp != nil {
				sendToConn(ws, outbound{Component: msg.Component, Payload: resp, Session: session.ID()})
				continue
			}
			if msg.Payload != nil && msg.Payload["init"] == true {
				sendToConn(ws, outbound{
					Component: msg.Component,
					Session:   session.ID(),
					Payload:   map[string]any{"session": session.ID()},
				})
			}
		}
	}
}

// Broadcast sends the given payload to all connections subscribed to the component name.
func Broadcast(name string, payload any, opts ...BroadcastOption) {
	var options broadcastOptions
	for _, opt := range opts {
		opt(&options)
	}

	connMu.RLock()
	conns := connections[name]
	connMu.RUnlock()
	if len(conns) == 0 {
		return
	}

	for ws, session := range conns {
		if options.session != "" && session.ID() != options.session {
			continue
		}
		sendToConn(ws, outbound{Component: name, Payload: payload, Session: session.ID()})
	}
}

func sendToConn(ws *websocket.Conn, out outbound) {
	b, err := json.Marshal(out)
	if err != nil {
		return
	}
	if err := websocket.Message.Send(ws, b); err != nil {
		log.Printf("send: %v", err)
	}
}
