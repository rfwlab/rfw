//go:build js && wasm

package hostclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	fncaching "github.com/mirkobrombin/go-foundation/v2/core/caching"
	fnres "github.com/mirkobrombin/go-foundation/v2/core/resiliency"

	dom "github.com/rfwlab/rfw/v2/dom"
	js "github.com/rfwlab/rfw/v2/js"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type componentBinding struct {
	id   string
	vars []string
}

var (
	conn      *websocket.Conn
	bindings  = map[string]componentBinding{}
	once      sync.Once
	mu        sync.RWMutex
	pending   []message
	outbox    = map[uint64]message{}
	handlers  = map[string]func(map[string]any){}
	dedup     = map[string]struct{}{}
	debug     bool
	cb        *fnres.CircuitBreaker
	sendCache *fncaching.InMemoryCache[string]
	hydrateCB *fnres.CircuitBreaker

	sessionMu sync.RWMutex
	sessionID string

	deliveryMu   sync.Mutex
	sendMu       sync.Mutex
	resumeToken  string
	nextOutbound uint64
	lastInbound  uint64

	callSequence atomic.Uint64
	callMu       sync.Mutex
	pendingCalls = map[string]chan actionReply{}
)

type message struct {
	name     string
	action   string
	id       string
	payload  any
	sequence uint64
}

type wireMessage struct {
	Component   string `json:"component,omitempty"`
	Action      string `json:"action,omitempty"`
	ID          string `json:"id,omitempty"`
	Payload     any    `json:"payload,omitempty"`
	Sequence    uint64 `json:"sequence"`
	Ack         uint64 `json:"ack,omitempty"`
	ResumeToken string `json:"resumeToken,omitempty"`
}

type messageWriter func(context.Context, *websocket.Conn, wireMessage) error

type actionReply struct {
	payload any
	err     *ActionError
}

func decodeInitSnapshotPayload(raw any) *initSnapshotPayload {
	if raw == nil {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	html, _ := m["html"].(string)
	if html == "" {
		return nil
	}
	var vars []string
	if list, ok := m["vars"].([]any); ok {
		vars = make([]string, 0, len(list))
		for _, item := range list {
			if s, ok := item.(string); ok {
				vars = append(vars, s)
			}
		}
	} else if list, ok := m["vars"].([]string); ok {
		vars = append(vars, list...)
	}
	return &initSnapshotPayload{HTML: html, Vars: vars}
}

// ActionError is a machine-readable error returned by a typed host action.
type ActionError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *ActionError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}

func init() {
	cb = fnres.NewCircuitBreaker(5, 30*time.Second)
	cb.OnStateChange(func(from, to fnres.State) {
		if debug {
			log.Printf("hostclient: circuit %v -> %v", from, to)
		}
	})
	hydrateCB = fnres.NewCircuitBreaker(3, 15*time.Second)
	sendCache = fncaching.NewInMemory[string](
		fncaching.WithMaxEntries[string](256),
		fncaching.WithTTL[string](5*time.Second),
	)
}

func connect() {
	once.Do(func() {
		go connectionLoop()
	})
}

// hostWSURL builds the WebSocket URL the client uses to reach its host.
// The endpoint is resolved in order of precedence: a full URL in
// window.RFW_HOST_URL (ws, wss, http, https, or a bare host[:port] with an
// optional path), the legacy host[:port] in window.RFW_HOST, or the page
// origin. The path defaults to /ws when the endpoint carries none.
func hostWSURL() string {
	if u := js.Get("RFW_HOST_URL"); u.Truthy() {
		if s := normalizeWSURL(u.String()); s != "" {
			return s
		}
	}
	host := js.Location().Get("host").String()
	if h := js.Get("RFW_HOST"); h.Truthy() {
		host = h.String()
	}
	return normalizeWSURL(host)
}

// normalizeWSURL turns a configured endpoint into a WebSocket URL: http and
// https map to ws and wss, a bare host takes the page scheme, and /ws is
// appended when the endpoint carries no path.
func normalizeWSURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(raw, "ws://"), strings.HasPrefix(raw, "wss://"):
	case strings.HasPrefix(raw, "http://"):
		raw = "ws://" + strings.TrimPrefix(raw, "http://")
	case strings.HasPrefix(raw, "https://"):
		raw = "wss://" + strings.TrimPrefix(raw, "https://")
	default:
		scheme := "wss"
		if js.Location().Get("protocol").String() == "http:" {
			scheme = "ws"
		}
		raw = scheme + "://" + raw
	}
	if rest := raw[strings.Index(raw, "://")+3:]; !strings.Contains(rest, "/") {
		raw = strings.TrimRight(raw, "/") + "/ws"
	}
	return raw
}

func connectionLoop() {
	for {
		url := hostWSURL()
		connectionState.Set(ConnectionConnecting)

		err := fnres.Retry(context.Background(), func() error {
			return cb.Execute(func() error {
				if debug {
					log.Printf("hostclient: dialing %s", url)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				c, _, derr := websocket.Dial(ctx, url, nil)
				if derr != nil {
					return derr
				}

				sendMu.Lock()
				mu.Lock()
				conn = c
				pend := pending
				pending = nil
				mu.Unlock()

				if debug {
					log.Printf("hostclient: connected")
				}
				connectionState.Set(ConnectionConnected)

				mu.RLock()
				names := make([]string, 0, len(bindings)+len(handlers))
				for name := range bindings {
					names = append(names, name)
				}
				for name := range handlers {
					if _, bound := bindings[name]; !bound {
						names = append(names, name)
					}
				}
				mu.RUnlock()
				deliveryMu.Lock()
				unacknowledged := make([]message, 0, len(outbox))
				for sequence := uint64(1); sequence <= nextOutbound; sequence++ {
					if msg, ok := outbox[sequence]; ok {
						unacknowledged = append(unacknowledged, msg)
					}
				}
				deliveryMu.Unlock()
				initialized := make(map[string]struct{})
				for _, msg := range unacknowledged {
					sendMessageUnlocked(c, msg)
					if name, ok := initMessageName(msg); ok {
						initialized[name] = struct{}{}
					}
				}
				for _, msg := range pend {
					sendMessageUnlocked(c, msg)
					if name, ok := initMessageName(msg); ok {
						initialized[name] = struct{}{}
					}
				}
				for _, name := range names {
					if _, sent := initialized[name]; sent {
						continue
					}
					sendMessageUnlocked(c, message{name: name, payload: map[string]any{"init": true}})
				}
				sendMu.Unlock()

				ctx2, cancel2 := context.WithCancel(context.Background())
				defer cancel2()
				errCh := make(chan error, 2)
				go func() { errCh <- readLoop(ctx2, c) }()
				go func() { errCh <- pingLoop(ctx2, c) }()
				loopErr := <-errCh
				cancel2()
				closeErr := c.Close(websocket.StatusInternalError, "connection closed")

				mu.Lock()
				conn = nil
				mu.Unlock()
				connectionState.Set(ConnectionDisconnected)
				if loopErr != nil {
					return loopErr
				}
				return closeErr
			})
		},
			fnres.WithAttempts(5),
			fnres.WithDelay(time.Second, 30*time.Second),
			fnres.WithFactor(2),
			fnres.WithJitter(0.1),
			fnres.WithRetryIf(func(err error) bool { return err != nil }),
		)

		if err != nil && debug {
			log.Printf("hostclient: connection attempt failed: %v", err)
		}
		connectionState.Set(ConnectionDisconnected)

		// Back off before reconnecting to avoid tight loops on persistent failures.
		time.Sleep(time.Second)
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
			Component   string       `json:"component"`
			Action      string       `json:"action"`
			Control     string       `json:"control"`
			ID          string       `json:"id"`
			Payload     any          `json:"payload"`
			Error       *ActionError `json:"error"`
			Session     string       `json:"session"`
			Sequence    uint64       `json:"sequence"`
			Ack         uint64       `json:"ack"`
			ResumeToken string       `json:"resumeToken"`
		}
		if err := wsjson.Read(ctx, c, &msg); err != nil {
			return err
		}
		if debug {
			log.Printf("hostclient: recv %s %v", msg.Component, msg.Payload)
		}
		prepareInboundDelivery(msg.Session, msg.Control)
		deliveryMu.Lock()
		for sequence := range outbox {
			if sequence <= msg.Ack {
				delete(outbox, sequence)
			}
		}
		if msg.Sequence != 0 {
			if msg.Sequence <= lastInbound {
				deliveryMu.Unlock()
				continue
			}
			if lastInbound != 0 && msg.Sequence != lastInbound+1 {
				deliveryMu.Unlock()
				connectionState.Set(ConnectionDesynced)
				return errors.New("hostclient: server message sequence gap")
			}
			lastInbound = msg.Sequence
		}
		if msg.ResumeToken != "" {
			resumeToken = msg.ResumeToken
		}
		deliveryMu.Unlock()
		if msg.ID != "" {
			callMu.Lock()
			replyChannel := pendingCalls[msg.ID]
			if replyChannel != nil {
				delete(pendingCalls, msg.ID)
			}
			callMu.Unlock()
			if replyChannel != nil {
				replyChannel <- actionReply{payload: msg.Payload, err: msg.Error}
				continue
			}
		}
		if msg.Control != "" {
			continue
		}
		payload, _ := msg.Payload.(map[string]any)
		if payload == nil {
			payload = make(map[string]any)
		}
		mu.RLock()
		h, hasHandler := handlers[msg.Component]
		b, hasBinding := bindings[msg.Component]
		mu.RUnlock()
		if hasHandler {
			if msg.Session != "" {
				payload["_session"] = msg.Session
			}
			h(payload)
			continue
		}
		if hasBinding {
			rootEl := dom.ComponentRoot(b.id)
			if !rootEl.Truthy() {
				continue
			}
			root := newComponentRoot(rootEl)

			if snap := decodeInitSnapshotPayload(payload["initSnapshot"]); snap != nil {
				applyInitSnapshot(root, snap)
				if len(snap.Vars) > 0 {
					b.vars = append([]string(nil), snap.Vars...)
					mu.Lock()
					bindings[msg.Component] = b
					mu.Unlock()
				}
				continue
			}

			mismatches := handleHostPayload(root, payload, func(name string, raw any) {
				sig := dom.SnapshotComponentSignals(b.id)
				if sig == nil {
					return
				}
				if s, ok := sig[name]; ok {
					if setter, ok := s.(interface{ SetFromHost(any) }); ok {
						setter.SetFromHost(raw)
					}
				}
			})
			if len(mismatches) > 0 {
				for _, m := range mismatches {
					log.Printf("hostclient: hydration mismatch component=%s var=%s expected=%s actualHash=%s actual=%q", msg.Component, m.VarName, m.Expected, m.ActualHash, m.Actual)
				}
				resyncErr := hydrateCB.Execute(func() error {
					Send(msg.Component, buildResyncPayload(mismatches))
					return nil
				})
				if resyncErr != nil {
					log.Printf("hostclient: hydration circuit open, skipping resync for %s", msg.Component)
				}
			}
		}
	}
}

func prepareInboundDelivery(remoteSession, control string) {
	sessionMu.Lock()
	previousSession := sessionID
	if remoteSession != "" {
		sessionID = remoteSession
	}
	sessionMu.Unlock()
	if control != "resume_rejected" && (remoteSession == "" || previousSession == "" || remoteSession == previousSession) {
		return
	}
	deliveryMu.Lock()
	lastInbound = 0
	resumeToken = ""
	deliveryMu.Unlock()
}

// RegisterComponent binds a client component to a host component name.
func RegisterComponent(id, name string, vars []string) {
	mu.Lock()
	bindings[name] = componentBinding{id: id, vars: vars}
	current := conn
	if current == nil {
		pending = append(pending, message{name: name, payload: map[string]any{"init": true}})
	}
	mu.Unlock()
	connect()
	if current != nil {
		sendMessage(current, message{name: name, payload: map[string]any{"init": true}})
	}
}

// EnableSendDedup turns on payload-based deduplication for the named channel:
// identical payloads sent within a 5 second window are dropped. Dedup is off
// by default because repeated identical messages are usually intentional user
// actions (e.g. clicking +1 twice); opt in only for channels where duplicate
// suppression is the desired semantic.
func EnableSendDedup(name string) {
	mu.Lock()
	dedup[name] = struct{}{}
	mu.Unlock()
}

func dedupEnabled(name string) bool {
	mu.RLock()
	_, ok := dedup[name]
	mu.RUnlock()
	return ok
}

// Send queues or transmits a host component message.
func Send(name string, payload any) {
	connect()
	if dedupEnabled(name) {
		key := fmt.Sprintf("%s|%v", name, payload)
		if _, ok, _ := sendCache.Get(context.Background(), key); ok {
			return
		}
		if err := sendCache.Set(context.Background(), key, "sent", 5*time.Second); err != nil {
			log.Printf("hostclient: dedup cache set failed: %v", err)
		}
	}

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

// RegisterHandler registers a handler for host messages.
func RegisterHandler(name string, h func(map[string]any)) {
	mu.Lock()
	handlers[name] = h
	current := conn
	if current == nil {
		pending = append(pending, message{name: name, payload: map[string]any{"init": true}})
	}
	mu.Unlock()
	connect()
	if current != nil {
		sendMessage(current, message{name: name, payload: map[string]any{"init": true}})
	}
}

// SessionID returns the current SSC session ID.
func SessionID() string {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return sessionID
}

func sendMessage(c *websocket.Conn, msg message) {
	sendMessageWithWriter(c, msg, writeMessage)
}

func sendMessageWithWriter(c *websocket.Conn, msg message, writer messageWriter) {
	sendMu.Lock()
	defer sendMu.Unlock()
	sendMessageUnlockedWithWriter(c, msg, writer)
}

func sendMessageUnlocked(c *websocket.Conn, msg message) {
	sendMessageUnlockedWithWriter(c, msg, writeMessage)
}

func sendMessageUnlockedWithWriter(c *websocket.Conn, msg message, writer messageWriter) {
	deliveryMu.Lock()
	if msg.sequence == 0 {
		nextOutbound++
		msg.sequence = nextOutbound
		outbox[msg.sequence] = msg
	}
	token := resumeToken
	ack := lastInbound
	deliveryMu.Unlock()
	outbound := wireMessage{
		Component:   msg.name,
		Action:      msg.action,
		ID:          msg.id,
		Payload:     msg.payload,
		Sequence:    msg.sequence,
		Ack:         ack,
		ResumeToken: token,
	}
	ctx := context.Background()
	_ = writer(ctx, c, outbound)
}

func writeMessage(ctx context.Context, c *websocket.Conn, message wireMessage) error {
	return wsjson.Write(ctx, c, message)
}

func initMessageName(msg message) (string, bool) {
	if msg.name == "" || msg.action != "" {
		return "", false
	}
	payload, ok := msg.payload.(map[string]any)
	if !ok || payload["init"] != true {
		return "", false
	}
	return msg.name, true
}

// Call invokes a typed SSC action and waits for its correlated response.
func Call[Request, Response any](ctx context.Context, action string, request Request) (Response, error) {
	var zero Response
	if ctx == nil {
		ctx = context.Background()
	}
	if action == "" {
		return zero, errors.New("hostclient: empty action name")
	}
	connect()
	id := fmt.Sprintf("call-%d", callSequence.Add(1))
	replyChannel := make(chan actionReply, 1)
	callMu.Lock()
	pendingCalls[id] = replyChannel
	callMu.Unlock()

	msg := message{action: action, id: id, payload: request}
	mu.RLock()
	current := conn
	mu.RUnlock()
	if current == nil {
		mu.Lock()
		pending = append(pending, msg)
		mu.Unlock()
	} else {
		sendMessage(current, msg)
	}

	select {
	case reply := <-replyChannel:
		if reply.err != nil {
			return zero, reply.err
		}
		data, err := json.Marshal(reply.payload)
		if err != nil {
			return zero, fmt.Errorf("hostclient: encode action response: %w", err)
		}
		if err := json.Unmarshal(data, &zero); err != nil {
			return zero, fmt.Errorf("hostclient: decode action response: %w", err)
		}
		return zero, nil
	case <-ctx.Done():
		callMu.Lock()
		delete(pendingCalls, id)
		callMu.Unlock()
		return zero, ctx.Err()
	}
}

// FormResponse is the typed result returned by host.RegisterForm.
type FormResponse[Response any] struct {
	Data   Response          `json:"data,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
	Valid  bool              `json:"valid"`
}

// SubmitForm invokes a typed SSC form action.
func SubmitForm[Values, Response any](ctx context.Context, action string, values Values) (FormResponse[Response], error) {
	return Call[Values, FormResponse[Response]](ctx, action, values)
}

// EnableDebug enables host client debug logging.
func EnableDebug() { debug = true }
