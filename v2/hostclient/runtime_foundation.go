//go:build js && wasm

package hostclient

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	fncaching "github.com/mirkobrombin/go-foundation/pkg/caching"
	fnres "github.com/mirkobrombin/go-foundation/pkg/resiliency"

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
	conn       *websocket.Conn
	bindings   = map[string]componentBinding{}
	once       sync.Once
	mu         sync.RWMutex
	pending    []message
	handlers   = map[string]func(map[string]any){}
	debug      bool
	cb         *fnres.CircuitBreaker
	sendCache  *fncaching.InMemoryCache[string]
	hydrateCB  *fnres.CircuitBreaker

	sessionMu sync.RWMutex
	sessionID string
)

type message struct {
	name    string
	payload any
}

func init() {
	cb = fnres.NewCircuitBreaker(5, 30*time.Second)
	cb.OnStateChange(func(from, to fnres.State) {
		if debug {
			log.Printf("hostclient: circuit %v → %v", from, to)
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

func connectionLoop() {
	host := js.Location().Get("host").String()
	if h := js.Get("RFW_HOST"); h.Truthy() {
		host = h.String()
	}
	scheme := "wss"
	if js.Location().Get("protocol").String() == "http:" {
		scheme = "ws"
	}
	for {
		url := fmt.Sprintf("%s://%s/ws", scheme, host)

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

				mu.Lock()
				conn = c
				pend := pending
				pending = nil
				mu.Unlock()

				if debug {
					log.Printf("hostclient: connected")
				}

				for _, msg := range pend {
					sendMessage(c, msg)
				}
				for name := range bindings {
					sendMessage(c, message{name: name, payload: map[string]any{"init": true}})
				}

				ctx2, cancel2 := context.WithCancel(context.Background())
				defer cancel2()
				errCh := make(chan error, 2)
				go func() { errCh <- readLoop(ctx2, c) }()
				go func() { errCh <- pingLoop(ctx2, c) }()
				loopErr := <-errCh
				cancel2()
				c.Close(websocket.StatusInternalError, "connection closed")

				mu.Lock()
				conn = nil
				mu.Unlock()
				return loopErr
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

			mismatches := handleHostPayload(root, msg.Payload, func(name string, raw any) {
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

func RegisterComponent(id, name string, vars []string) {
	mu.Lock()
	bindings[name] = componentBinding{id: id, vars: vars}
	mu.Unlock()
	connect()
}

func Send(name string, payload any) {
	connect()
	key := fmt.Sprintf("%v", payload)
	if _, ok, _ := sendCache.Get(context.Background(), key); ok {
		return
	}
	sendCache.Set(context.Background(), key, "sent", 5*time.Second)

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

func RegisterHandler(name string, h func(map[string]any)) {
	mu.Lock()
	handlers[name] = h
	mu.Unlock()
	connect()
	Send(name, map[string]any{"init": true})
}

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

func EnableDebug() { debug = true }