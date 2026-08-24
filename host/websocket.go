package host

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"

	"golang.org/x/net/websocket"
)

// BroadcastOption configures a broadcast call.
type BroadcastOption func(*BroadcastOptions)

// BroadcastOptions holds optional parameters for Broadcast.
type BroadcastOptions struct {
	Session string
}

// WithSessionTarget limits a broadcast to a specific session ID.
func WithSessionTarget(sessionID string) BroadcastOption {
	return func(opts *BroadcastOptions) {
		opts.Session = sessionID
	}
}

var (
	connections = make(map[string]map[*websocket.Conn]*Session)
	connMu      sync.RWMutex
	connWrites  sync.Map
)

// AnswerControl replies to an out-of-band control frame and reports whether it
// consumed the message.
//
// A browser client cannot send protocol ping frames, so liveness rides on a
// control message instead. It carries no sequence, so its answer stays out of
// the replay history and out of the message budget, matching the frame level
// pong it replaces.
//
// Both this package's handler and the one in ssc route inbound frames through
// here. They are separate loops over the same protocol, and when only one of
// them answered a ping an idle connection to an SSC server dropped and
// reconnected on every heartbeat.
func AnswerControl(ws *websocket.Conn, msg Inbound) bool {
	if msg.Control != "ping" {
		return false
	}
	SendOutbound(ws, Outbound{Control: "pong"})
	return true
}

func wsHandler(ws *websocket.Conn, runtime *WSRuntime) {
	if !runtime.AcquireConnection() {
		SendOutbound(ws, Outbound{Error: NewActionError("connection_limit", "connection limit reached")})
		if err := ws.Close(); err != nil {
			log.Printf("close rejected websocket: %v", err)
		}
		return
	}
	defer runtime.ReleaseConnection()
	runtime.ConfigureConnection(ws)

	var session *Session
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
		SuspendSession(session, runtime.ResumeTTL())
		ForgetConnection(ws)
		if err := ws.Close(); err != nil {
			log.Printf("close websocket: %v", err)
		}
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
		var msg Inbound
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("unmarshal: %v", err)
			continue
		}
		if session == nil {
			var resumed bool
			var err error
			session, resumed, err = runtime.OpenSession(ws.Request(), msg.ResumeToken)
			if err != nil {
				SendOutbound(ws, Outbound{Error: NewActionError("session_rejected", "session rejected")})
				return
			}
			BindSessionConnection(ws, session)
			if resumed {
				ReplaySession(ws, session, msg.Ack)
			} else if msg.ResumeToken != "" {
				SendSessionOutbound(ws, session, Outbound{
					Control: "resume_rejected",
					Error:   NewActionError("resume_rejected", "session could not be resumed"),
				})
			}
		}
		session.Acknowledge(msg.Ack)
		if AnswerControl(ws, msg) {
			continue
		}
		if err := session.AcceptInbound(msg.Sequence); err != nil {
			if errors.Is(err, ErrDuplicateMessage) {
				continue
			}
			SendSessionOutbound(ws, session, Outbound{
				ID:     msg.ID,
				Action: msg.Action,
				Error:  NewActionError("sequence_gap", "client message sequence gap"),
			})
			continue
		}
		if !session.AllowMessage(runtime.MessagesPerMinute()) {
			SendSessionOutbound(ws, session, Outbound{
				ID:     msg.ID,
				Action: msg.Action,
				Error:  NewActionError("rate_limited", "message rate limit exceeded"),
			})
			continue
		}
		authorizeCtx, cancelAuthorize := runtime.HandlerContext(context.Background())
		authorizeErr := runtime.Authorize(authorizeCtx, session, msg)
		cancelAuthorize()
		if authorizeErr != nil {
			SendSessionOutbound(ws, session, Outbound{
				Component: msg.Component,
				Action:    msg.Action,
				ID:        msg.ID,
				Error:     NewActionError("forbidden", "message forbidden"),
			})
			continue
		}
		if msg.Action != "" {
			payload, actionErr := runtime.DispatchAction(context.Background(), session, msg)
			SendSessionOutbound(ws, session, Outbound{
				Action:  msg.Action,
				ID:      msg.ID,
				Payload: payload,
				Error:   actionErr,
			})
			continue
		}
		if msg.Component != "" && msg.Payload != nil && msg.Payload["unsubscribe"] == true {
			connMu.Lock()
			if set, ok := connections[msg.Component]; ok {
				delete(set, ws)
				if len(set) == 0 {
					delete(connections, msg.Component)
				}
			}
			for index, name := range subscribed {
				if name == msg.Component {
					subscribed = append(subscribed[:index], subscribed[index+1:]...)
					break
				}
			}
			connMu.Unlock()
			SendSessionOutbound(ws, session, Outbound{Component: msg.Component, Control: "unsubscribed"})
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
				switch v := resp.(type) {
				case *InitSnapshot:
					if v != nil {
						SendSessionOutbound(ws, session, Outbound{Component: msg.Component, ID: msg.ID, Payload: map[string]any{"initSnapshot": v}})
					}
					continue
				case InitSnapshot:
					SendSessionOutbound(ws, session, Outbound{Component: msg.Component, ID: msg.ID, Payload: map[string]any{"initSnapshot": v}})
					continue
				default:
					SendSessionOutbound(ws, session, Outbound{Component: msg.Component, ID: msg.ID, Payload: resp})
					continue
				}
			}
			if msg.Payload != nil && msg.Payload["init"] == true {
				SendSessionOutbound(ws, session, Outbound{
					Component: msg.Component,
					Payload:   map[string]any{"session": session.ID()},
				})
				continue
			}
		}
		SendSessionOutbound(ws, session, Outbound{Control: "ack"})
	}
}

// Broadcast sends the given payload to all connections subscribed to the component name.
func Broadcast(name string, payload any, opts ...BroadcastOption) {
	var options BroadcastOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Snapshot the (conn, session) pairs under the lock: wsHandler mutates the
	// connection map on subscribe/disconnect, so iterating it after releasing
	// connMu races with those writes. Sends happen outside the lock.
	type target struct {
		ws      *websocket.Conn
		session *Session
	}
	connMu.RLock()
	targets := make([]target, 0, len(connections[name]))
	for ws, session := range connections[name] {
		targets = append(targets, target{ws: ws, session: session})
	}
	connMu.RUnlock()

	for _, t := range targets {
		if options.Session != "" && t.session.ID() != options.Session {
			continue
		}
		SendSessionOutbound(t.ws, t.session, Outbound{Component: name, Payload: payload})
	}
}

// DispatchAction executes a typed action within the configured handler deadline.
func (runtime *WSRuntime) DispatchAction(parent context.Context, session *Session, message Inbound) (any, *ActionError) {
	ctx, cancel := runtime.HandlerContext(parent)
	defer cancel()
	type result struct {
		payload any
		err     *ActionError
	}
	resultChannel := make(chan result, 1)
	go func() {
		defer func() {
			if recover() != nil {
				resultChannel <- result{err: NewActionError("action_failed", "action failed")}
			}
		}()
		payload, actionErr := DispatchAction(ctx, session, message.Action, message.Payload)
		resultChannel <- result{payload: payload, err: actionErr}
	}()
	select {
	case response := <-resultChannel:
		return response.payload, response.err
	case <-ctx.Done():
		return nil, NewActionError("action_timeout", "action timed out")
	}
}

// ReplaySession sends retained messages after the client's acknowledgement.
func ReplaySession(ws *websocket.Conn, session *Session, acknowledged uint64) {
	if session == nil {
		return
	}
	session.outboundMu.Lock()
	defer session.outboundMu.Unlock()
	accepted, _ := sessionAcceptsConnection(session, ws, true)
	if !accepted {
		return
	}
	lock := connectionWriteLock(ws)
	lock.Lock()
	defer lock.Unlock()
	messages, err := session.ReplayAfter(acknowledged)
	if err != nil {
		sendOutboundUnlocked(ws, session.PrepareOutbound(Outbound{
			Error: NewActionError("resync_required", "message history is no longer available"),
		}))
		return
	}
	for _, message := range messages {
		sendOutboundUnlocked(ws, message)
	}
}

// SendSessionOutbound assigns delivery metadata and sends a message. It binds
// ws when the session has no active connection. After ResumeSession, callers
// must complete the handoff with ReplaySession or BindSessionConnection before
// sending.
func SendSessionOutbound(ws *websocket.Conn, session *Session, out Outbound) {
	if session == nil {
		return
	}
	session.outboundMu.Lock()
	defer session.outboundMu.Unlock()
	accepted, handoffPending := sessionAcceptsConnection(session, ws, false)
	if !accepted {
		if handoffPending {
			logger.Debug("session outbound dropped before connection handoff", "session", session.ID())
		}
		return
	}
	lock := connectionWriteLock(ws)
	lock.Lock()
	defer lock.Unlock()
	sendOutboundUnlocked(ws, session.PrepareOutbound(out))
}

// BindSessionConnection marks ws as the active connection for session delivery.
// It binds a session without an active connection and completes the connection
// handoff after ResumeSession. It does not replace an active connection.
func BindSessionConnection(ws *websocket.Conn, session *Session) {
	if session == nil {
		return
	}
	session.outboundMu.Lock()
	accepted, _ := sessionAcceptsConnection(session, ws, true)
	if accepted {
		session.connectionManaged = true
	}
	session.outboundMu.Unlock()
}

func sessionAcceptsConnection(session *Session, ws *websocket.Conn, handoff bool) (bool, bool) {
	if ws == nil {
		return false, false
	}
	session.deliveryMu.Lock()
	active := session.attached && !session.released
	resumePending := session.resumePending
	if active && resumePending && handoff {
		session.resumePending = false
	}
	session.deliveryMu.Unlock()
	if !active {
		return false, false
	}
	if resumePending && !handoff {
		return false, true
	}
	if session.connection == ws {
		return true, false
	}
	if session.connection != nil || (session.connectionManaged && !handoff) {
		return false, false
	}
	session.connection = ws
	return true, false
}

// SendOutbound serializes writes per connection.
func SendOutbound(ws *websocket.Conn, out Outbound) {
	lock := connectionWriteLock(ws)
	lock.Lock()
	defer lock.Unlock()
	sendOutboundUnlocked(ws, out)
}

func connectionWriteLock(ws *websocket.Conn) *sync.Mutex {
	lockValue, _ := connWrites.LoadOrStore(ws, &sync.Mutex{})
	return lockValue.(*sync.Mutex)
}

func sendOutboundUnlocked(ws *websocket.Conn, out Outbound) {
	b, err := json.Marshal(out)
	if err != nil {
		return
	}
	if err := websocket.Message.Send(ws, b); err != nil {
		log.Printf("send: %v", err)
	}
}

// ForgetConnection releases the connection write lock.
func ForgetConnection(ws *websocket.Conn) {
	connWrites.Delete(ws)
}
