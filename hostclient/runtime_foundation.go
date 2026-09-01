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
)

type componentBinding struct {
	id   string
	vars []string
	// gate carries the registration's ownership into the host signal writes a
	// frame makes, which happen with no binding lock held.
	gate *deliveryGate
}

// deliveryGate is the revocation switch a registration hands to every frame
// delivered under it. A release closes it inside the same section that drops
// the registration and then waits out the writes it may already have allowed,
// so a frame the read loop snapshotted before a cleanup cannot update a host
// signal once that cleanup returned. Reading it costs one atomic load and takes
// no lock, so a signal setter that re-enters registration or release cannot
// deadlock against a delivery holding it.
type deliveryGate struct{ closed atomic.Bool }

// open reports whether the registration this gate belongs to is still the live
// one. The zero binding carries no gate and owns nothing.
func (g *deliveryGate) open() bool { return g != nil && !g.closed.Load() }

func (g *deliveryGate) close() {
	if g != nil {
		g.closed.Store(true)
	}
}

// gatedHostSetter is a host signal that can fold the caller's ownership check
// into its own store, so the two are one step against a concurrent release.
type gatedHostSetter interface {
	SetFromHostGated(raw any, allow func() bool) bool
}

// hostSetter is the plain host signal contract, without the gate.
type hostSetter interface{ SetFromHost(any) }

// hostWriteBarrier waits out a gated write that was already allowed.
type hostWriteBarrier interface{ HostWriteBarrier() }

var (
	conn                 *hostConn
	bindings             = map[string]componentBinding{}
	bindingTokens        = map[string]uint64{}
	once                 sync.Once
	mu                   sync.RWMutex
	pending              []message
	outbox               = map[uint64]message{}
	handlers             = map[string]func(map[string]any){}
	handlerTokens        = map[string]uint64{}
	dedup                = map[string]struct{}{}
	debug                bool
	cb                   *fnres.CircuitBreaker
	sendCache            *fncaching.InMemoryCache[string]
	hydrateCB            *fnres.CircuitBreaker
	lifecycleMu          sync.RWMutex
	connectionGeneration atomic.Uint64

	sessionMu sync.RWMutex
	sessionID string

	deliveryMu   sync.Mutex
	sendMu       sync.Mutex
	resumeToken  string
	nextOutbound uint64
	lastInbound  uint64

	callSequence    atomic.Uint64
	handlerSequence atomic.Uint64
	bindingSequence atomic.Uint64
	callMu          sync.Mutex
	pendingCalls    = map[string]chan actionReply{}

	// bindingMu serializes host binding delivery against registration and
	// release. A frame is validated against the live registration token and its
	// DOM applied while it is held, so a release that returns has already waited
	// out any in-flight DOM update and every later frame fails the token check.
	// It is always taken before mu and never while mu is held. Host signal
	// writes run application code and cannot be held under it; they carry the
	// registration's deliveryGate instead, and a signal lock is never taken
	// while either binding lock is held.
	bindingMu sync.Mutex
	// afterBindingSnapshot is nil outside tests and is read under mu with the
	// binding snapshot itself. It opens the window between snapshotting a
	// binding in the read loop and taking bindingMu, which is where a
	// concurrent release has to win.
	afterBindingSnapshot func(component string)
	// beforeHostSignalWrite is nil outside tests and is read under mu. It opens
	// the window between the ownership check a delivery makes and the write
	// itself, the one the gate exists to close.
	beforeHostSignalWrite func(component, name string)
)

// Host snapshots can legitimately reach several megabytes (for example, a
// paginated table's first hydration). Keep a finite ceiling so malformed peers
// still cannot grow memory without bound.
const maxInboundMessageBytes int64 = 8 << 20

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
	Control     string `json:"control,omitempty"`
	ID          string `json:"id,omitempty"`
	Payload     any    `json:"payload,omitempty"`
	Sequence    uint64 `json:"sequence"`
	Ack         uint64 `json:"ack,omitempty"`
	ResumeToken string `json:"resumeToken,omitempty"`
}

type messageWriter func(context.Context, *hostConn, wireMessage) error

type actionReply struct {
	payload  any
	err      *ActionError
	resetErr error
}

// ErrSessionReset is returned by an in-flight Call when ResetSession discards
// the authenticated SSC session that owned it.
var ErrSessionReset = errors.New("hostclient: session reset")

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
		go func() {
			for {
				js.Guard("host connection loop", connectionLoop)
				time.Sleep(time.Second)
			}
		}()
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
	// An endpoint that carries no path gets the default one. A bare trailing
	// slash is no path either: "https://host/" would otherwise dial the root,
	// which the host does not serve.
	rest := raw[strings.Index(raw, "://")+3:]
	if slash := strings.Index(rest, "/"); slash == -1 || strings.Trim(rest[slash:], "/") == "" {
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
				generation := connectionGeneration.Load()
				if debug {
					log.Printf("hostclient: dialing %s", url)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				c, derr := dial(ctx, url)
				if derr != nil {
					return derr
				}
				c.generation = generation
				lifecycleMu.RLock()
				if generation != connectionGeneration.Load() {
					lifecycleMu.RUnlock()
					_ = c.close()
					return nil
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
				lifecycleMu.RUnlock()

				ctx2, cancel2 := context.WithCancel(context.Background())
				defer cancel2()
				errCh := make(chan error, 2)
				go func() { errCh <- guardedLoop("host read loop", func() error { return readLoop(ctx2, c) }) }()
				go func() { errCh <- guardedLoop("host heartbeat loop", func() error { return heartbeatLoop(ctx2, c) }) }()
				loopErr := <-errCh
				cancel2()
				closeErr := c.close()

				mu.Lock()
				if conn == c {
					conn = nil
				}
				mu.Unlock()
				connectionState.Set(ConnectionDisconnected)
				if generation != connectionGeneration.Load() {
					return nil
				}
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

func guardedLoop(context string, fn func() error) error {
	var err error
	if !js.Guard(context, func() { err = fn() }) {
		return fmt.Errorf("%s panicked", context)
	}
	return err
}

func heartbeatLoop(ctx context.Context, c *hostConn) error {
	return c.heartbeat(ctx, heartbeatInterval, heartbeatTimeout)
}

func readLoop(ctx context.Context, c *hostConn) error {
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
		frame, err := c.read(ctx)
		if err != nil {
			return err
		}
		if c.generation != connectionGeneration.Load() {
			return ErrSessionReset
		}
		if err := json.Unmarshal(frame, &msg); err != nil {
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
		_, hasBinding := bindings[msg.Component]
		token := bindingTokens[msg.Component]
		barrier := afterBindingSnapshot
		mu.RUnlock()
		if hasHandler {
			if msg.Session != "" {
				payload["_session"] = msg.Session
			}
			js.Guard("host handler: "+msg.Component, func() { h(payload) })
			continue
		}
		if hasBinding {
			if barrier != nil {
				barrier(msg.Component)
			}
			js.Guard("host binding: "+msg.Component, func() {
				applyHostBinding(msg.Component, payload, token)
			})
		}
	}
}

// hostSignalUpdate is one host variable a delivered frame carries, held until
// the DOM work is done and its signal can be set outside bindingMu.
type hostSignalUpdate struct {
	name  string
	value any
}

// applyHostBinding delivers one host frame to the binding that owns it: the DOM
// of the root under bindingMu, then the signals and any resync with the lock
// released, since both run code that reacquires it.
func applyHostBinding(component string, payload map[string]any, token uint64) {
	updates, mismatches := deliverHostFrame(component, payload, token)
	applyHostSignals(component, token, updates)
	if len(mismatches) == 0 {
		return
	}
	for _, mismatch := range mismatches {
		log.Printf("hostclient: hydration mismatch component=%s var=%s expected=%s actualHash=%s actual=%q", component, mismatch.VarName, mismatch.Expected, mismatch.ActualHash, mismatch.Actual)
	}
	resyncErr := hydrateCB.Execute(func() error {
		Send(component, buildResyncPayload(mismatches))
		return nil
	})
	if resyncErr != nil {
		log.Printf("hostclient: hydration circuit open, skipping resync for %s", component)
	}
}

// deliverHostFrame updates the DOM of the root the binding owns and reports the
// signal updates the frame carries. It runs under bindingMu and re-reads the
// registration token there: a frame the read loop snapshotted before a cleanup
// blocks until the release completes and then finds the token gone, so it
// reaches neither the released root nor the replacement registered under the
// same name. mu is taken only for the token read and the snapshot bookkeeping.
func deliverHostFrame(component string, payload map[string]any, token uint64) ([]hostSignalUpdate, []hydrationMismatch) {
	bindingMu.Lock()
	defer bindingMu.Unlock()

	binding, live := liveBinding(component, token)
	if !live {
		return nil, nil
	}
	rootEl := hostComponentRoot(binding.id)
	if !rootEl.Truthy() {
		return nil, nil
	}
	root := newComponentRoot(rootEl)
	if snap := decodeInitSnapshotPayload(payload["initSnapshot"]); snap != nil {
		applyInitSnapshot(root, snap)
		if len(snap.Vars) > 0 {
			binding.vars = append([]string(nil), snap.Vars...)
			mu.Lock()
			// Only the registration this frame was validated against may be
			// updated: a snapshot must not reinstate a released binding nor
			// overwrite a newer one.
			if bindingTokens[component] == token {
				bindings[component] = binding
			}
			mu.Unlock()
		}
		return nil, nil
	}

	var updates []hostSignalUpdate
	mismatches := handleHostPayload(root, payload, func(name string, raw any) {
		updates = append(updates, hostSignalUpdate{name: name, value: raw})
	})
	return updates, mismatches
}

// applyHostSignals pushes a delivered frame into the component's host signals.
// A setter runs application code, which may register or release a host binding
// in turn, so it must not run under bindingMu. The registration's gate carries
// the ownership check into the write instead: the signal evaluates it under the
// same lock it stores the value with, so a release either refused the write or
// waited it out before returning, and neither side holds a binding lock while
// application code runs.
func applyHostSignals(component string, token uint64, updates []hostSignalUpdate) {
	if len(updates) == 0 {
		return
	}
	binding, live := liveBinding(component, token)
	if !live {
		return
	}
	signals := dom.SnapshotComponentSignals(binding.id)
	if len(signals) == 0 {
		return
	}
	hook := hostSignalWriteHook()
	for _, update := range updates {
		signal, ok := signals[update.name]
		if !ok {
			continue
		}
		if hook != nil {
			hook(component, update.name)
		}
		if !applyHostSignal(signal, binding.gate, update.value) {
			// The registration was released mid-frame: the rest of the frame
			// was addressed to it too.
			return
		}
	}
}

// applyHostSignal writes one update and reports whether the registration was
// still live for it. A signal that predates the gated setter cannot fold the
// check into its store, so it is checked before the write instead, which leaves
// the window the gate exists to close.
func applyHostSignal(signal any, gate *deliveryGate, value any) bool {
	if gated, ok := signal.(gatedHostSetter); ok {
		if gated.SetFromHostGated(value, gate.open) {
			return true
		}
		// A refused write is either a revoked registration or a payload the
		// signal cannot represent; only the first one ends the frame.
		return gate.open()
	}
	setter, ok := signal.(hostSetter)
	if !ok {
		return true
	}
	if !gate.open() {
		return false
	}
	setter.SetFromHost(value)
	return true
}

// fenceHostSignalWrites waits out the host signal writes a now closed gate had
// already allowed. Each barrier takes only that signal's own value lock, which
// never covers application code, so a release re-entered from a setter never
// waits on itself. It covers the signals the component still has registered,
// which is why core unmounts a component by releasing its host bindings before
// it drops its signals.
func fenceHostSignalWrites(id string) {
	if id == "" {
		return
	}
	for _, signal := range dom.SnapshotComponentSignals(id) {
		if barrier, ok := signal.(hostWriteBarrier); ok {
			barrier.HostWriteBarrier()
		}
	}
}

func hostSignalWriteHook() func(component, name string) {
	mu.RLock()
	defer mu.RUnlock()
	return beforeHostSignalWrite
}

// liveBinding returns the binding still registered under token. A token that no
// longer matches means the registration was released or replaced, so the caller
// owns nothing to update.
func liveBinding(component string, token uint64) (componentBinding, bool) {
	mu.RLock()
	defer mu.RUnlock()
	if bindingTokens[component] != token {
		return componentBinding{}, false
	}
	binding, bound := bindings[component]
	return binding, bound
}

// hostComponentRoot resolves the exact root a binding owns. dom.ComponentRoot
// falls back to #app when the id matches nothing, which for host delivery would
// mean a frame addressed to an unmounted component writing into the application
// shell or into whatever mounted after it. A missing root is a frame to ignore.
// RegisterComponent is public and takes any id, so the id never reaches a
// selector unescaped: a quote or a bracket in it would otherwise throw a
// DOMException out of querySelector, or select a root the caller never owned.
func hostComponentRoot(id string) dom.Element {
	if id == "" {
		return dom.Element{Value: js.Null()}
	}
	if selector, ok := componentIDSelector(id); ok {
		return dom.Doc().Query(selector)
	}
	return scanComponentRoot(id)
}

// componentIDSelector builds the attribute selector for id and leaves the
// escaping rules to the browser. CSS.escape returns an identifier, which is
// what the value position of an attribute selector accepts.
func componentIDSelector(id string) (string, bool) {
	css := js.Get("CSS")
	if !css.Truthy() || css.Get("escape").Type() != js.TypeFunction {
		return "", false
	}
	return "[data-component-id=" + css.Call("escape", id).String() + "]", true
}

// scanComponentRoot is the fallback for a runtime without CSS.escape: the
// candidates are selected on the bare attribute and compared on its value, so
// no part of the id is ever parsed as a selector.
func scanComponentRoot(id string) dom.Element {
	roots := dom.Doc().QueryAll("[data-component-id]")
	if !roots.Truthy() {
		return dom.Element{Value: js.Null()}
	}
	for index := 0; index < roots.Length(); index++ {
		candidate := roots.Index(index)
		if candidate.Attr("data-component-id") == id {
			return candidate
		}
	}
	return dom.Element{Value: js.Null()}
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

// RegisterComponent binds a client component to a host component name. The
// binding lives until another registration replaces it. Use
// RegisterComponentOwned when the caller has to release the binding again, for
// instance because its component can unmount.
func RegisterComponent(id, name string, vars []string) {
	registerComponent(id, name, vars)
}

// RegisterComponentOwned binds like RegisterComponent and returns an idempotent
// cleanup that owns the binding's lifecycle. Releasing removes the binding from
// reconnect hydration and tells the active host session to stop broadcasts for
// the component; once the cleanup returns, no frame still in flight and no
// later frame can update the root it owned or write into its host signals. A
// stale cleanup closure never removes a newer binding registered under the same
// name. The cleanup is safe to call from a host signal setter, which is where a
// component that unmounts on an update ends up calling it.
func RegisterComponentOwned(id, name string, vars []string) func() {
	token := registerComponent(id, name, vars)
	var once sync.Once
	return func() {
		once.Do(func() { releaseComponent(name, token) })
	}
}

// registerComponent installs the binding and returns the token identifying this
// registration. Delivery validates against it, so registering also invalidates
// the cleanup of the binding it replaced.
func registerComponent(id, name string, vars []string) uint64 {
	token := bindingSequence.Add(1)
	bindingMu.Lock()
	mu.Lock()
	previous := bindings[name]
	bindings[name] = componentBinding{id: id, vars: vars, gate: &deliveryGate{}}
	bindingTokens[name] = token
	current := conn
	recordPendingControl(name, map[string]any{"init": true}, current)
	mu.Unlock()
	// The registration this one replaces stops being deliverable here. Its
	// frames already fail the token check; closing its gate stops one that was
	// snapshotted before the replacement from writing a signal on its way out.
	previous.gate.close()
	bindingMu.Unlock()
	connect()
	if current != nil {
		sendMessage(current, message{name: name, payload: map[string]any{"init": true}})
	}
	return token
}

func releaseComponent(name string, token uint64) {
	bindingMu.Lock()
	mu.Lock()
	if bindingTokens[name] != token {
		mu.Unlock()
		bindingMu.Unlock()
		return
	}
	binding := bindings[name]
	delete(bindings, name)
	delete(bindingTokens, name)
	current := conn
	recordPendingControl(name, map[string]any{"unsubscribe": true}, current)
	mu.Unlock()
	// Closing the gate in the section that dropped the registration means no
	// delivery can read it open again afterwards.
	binding.gate.close()
	// Every delivery for this binding is now either finished or bound to fail
	// its token check, so the unsubscribe can go out without the lock.
	bindingMu.Unlock()

	// The root is safe by now, the signals are not: a frame that read the gate
	// open before it closed may be committing a write. Waiting that write out,
	// with no binding lock held, is what makes the cleanup final.
	fenceHostSignalWrites(binding.id)

	if current != nil {
		sendMessage(current, message{name: name, payload: map[string]any{"unsubscribe": true}})
	}
}

// recordPendingControl keeps the reconnect queue holding one registration
// control per host component, the latest desired state: a registration
// supersedes a queued unsubscribe and a release supersedes a queued init, so
// route churn while offline cannot retain one message per cycle. The control
// the caller just decided on is queued only when there is no connection to send
// it on; a superseded one is dropped either way, since sending the new state
// makes replaying the old one wrong.
//
// The new state takes the queue position of the control it supersedes rather
// than the tail: everything else in the queue, the controls of other names and
// this name's own messages, keeps the order it was queued in, so a reconnect
// replays the component's init before the command that assumed it.
// It must be called with mu held.
func recordPendingControl(name string, payload map[string]any, current *hostConn) {
	replaced := false
	filtered := pending[:0]
	for _, queued := range pending {
		if queued.name != name || !isRegistrationControl(queued.payload) {
			filtered = append(filtered, queued)
			continue
		}
		if current != nil || replaced {
			continue
		}
		queued.payload = payload
		filtered = append(filtered, queued)
		replaced = true
	}
	pending = filtered
	if current == nil && !replaced {
		pending = append(pending, message{name: name, payload: payload})
	}
}

func isRegistrationControl(payload any) bool {
	values, ok := payload.(map[string]any)
	if !ok {
		return false
	}
	return values["init"] == true || values["unsubscribe"] == true
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
	lifecycleMu.RLock()
	defer lifecycleMu.RUnlock()
	connect()
	if dedupEnabled(name) {
		key := fmt.Sprintf("%d|%s|%v", connectionGeneration.Load(), name, payload)
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

// RegisterHandler registers a handler for host messages and returns an
// idempotent unsubscribe function. Unsubscribing removes the handler from
// reconnect hydration and tells the active host session to stop broadcasts for
// the component. A stale unsubscribe closure never removes a newer handler
// registered under the same name.
func RegisterHandler(name string, h func(map[string]any)) func() {
	token := handlerSequence.Add(1)
	mu.Lock()
	handlers[name] = h
	handlerTokens[name] = token
	current := conn
	recordPendingControl(name, map[string]any{"init": true}, current)
	mu.Unlock()
	connect()
	if current != nil {
		sendMessage(current, message{name: name, payload: map[string]any{"init": true}})
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			mu.Lock()
			if handlerTokens[name] != token {
				mu.Unlock()
				return
			}
			delete(handlers, name)
			delete(handlerTokens, name)
			current := conn
			recordPendingControl(name, map[string]any{"unsubscribe": true}, current)
			mu.Unlock()

			if current != nil {
				sendMessage(current, message{name: name, payload: map[string]any{"unsubscribe": true}})
			}
		})
	}
}

// SessionID returns the current SSC session ID.
func SessionID() string {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return sessionID
}

// ResetSession closes the active SSC transport and discards all delivery
// state owned by its authenticated session. Live registrations are preserved;
// callers must unmount user-owned scopes before invoking it.
func ResetSession() {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	connectionGeneration.Add(1)
	connectionState.Set(ConnectionDisconnected)
	mu.Lock()
	current := conn
	conn = nil
	pending = nil
	mu.Unlock()
	sessionMu.Lock()
	sessionID = ""
	sessionMu.Unlock()
	deliveryMu.Lock()
	resumeToken = ""
	nextOutbound = 0
	lastInbound = 0
	outbox = map[uint64]message{}
	deliveryMu.Unlock()
	callMu.Lock()
	calls := pendingCalls
	pendingCalls = map[string]chan actionReply{}
	callMu.Unlock()
	for _, reply := range calls {
		reply <- actionReply{resetErr: ErrSessionReset}
	}
	if current != nil {
		_ = current.close()
	}
}

func sendMessage(c *hostConn, msg message) {
	sendMessageWithWriter(c, msg, writeMessage)
}

func sendMessageWithWriter(c *hostConn, msg message, writer messageWriter) {
	sendMu.Lock()
	defer sendMu.Unlock()
	sendMessageUnlockedWithWriter(c, msg, writer)
}

func sendMessageUnlocked(c *hostConn, msg message) {
	sendMessageUnlockedWithWriter(c, msg, writeMessage)
}

func sendMessageUnlockedWithWriter(c *hostConn, msg message, writer messageWriter) {
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

func writeMessage(_ context.Context, c *hostConn, message wireMessage) error {
	return c.writeJSON(message)
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
	lifecycleMu.RLock()
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
	lifecycleMu.RUnlock()

	select {
	case reply := <-replyChannel:
		if reply.resetErr != nil {
			return zero, reply.resetErr
		}
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
