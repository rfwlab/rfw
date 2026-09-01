//go:build js && wasm

package hostclient

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

var runtimeTestSequence atomic.Uint64

func uniqueRuntimeTestName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, runtimeTestSequence.Add(1))
}

func TestGuardedLoopConvertsPanicAndNextLoopRuns(t *testing.T) {
	previous := js.OnRuntimePanic
	defer func() { js.OnRuntimePanic = previous }()
	recovered := 0
	js.OnRuntimePanic = func(any, string, []byte) { recovered++ }

	if err := guardedLoop("read", func() error { panic("bad push") }); err == nil {
		t.Fatal("panicking loop returned nil")
	}
	if err := guardedLoop("read", func() error { return nil }); err != nil {
		t.Fatalf("next loop returned %v", err)
	}
	if recovered != 1 {
		t.Fatalf("recovered loop panics = %d, want 1", recovered)
	}
}

func TestInboundMessageLimitSupportsHydrationSnapshots(t *testing.T) {
	if maxInboundMessageBytes < 1<<20 {
		t.Fatalf("inbound message limit = %d, want at least 1 MiB", maxInboundMessageBytes)
	}
	if maxInboundMessageBytes > 16<<20 {
		t.Fatalf("inbound message limit = %d, want a bounded ceiling", maxInboundMessageBytes)
	}
}

func pendingCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(pending)
}

func TestRegisterHandlerUnsubscribeQueuesWireUnsubscribe(t *testing.T) {
	name := uniqueRuntimeTestName("scoped-handler")
	before := pendingCount()
	unsubscribe := RegisterHandler(name, func(map[string]any) {})
	if got := pendingCount() - before; got != 1 {
		t.Fatalf("queued subscribe messages = %d, want 1", got)
	}
	unsubscribe()

	mu.RLock()
	_, stillRegistered := handlers[name]
	queued := append([]message(nil), pending...)
	mu.RUnlock()
	if stillRegistered {
		t.Fatal("handler remained registered after unsubscribe")
	}
	count := 0
	for _, item := range queued {
		if item.name == name {
			values, _ := item.payload.(map[string]any)
			if values["unsubscribe"] == true {
				count++
			}
			if values["init"] == true {
				t.Fatal("stale init remained queued after unsubscribe")
			}
		}
	}
	if count != 1 {
		t.Fatalf("queued unsubscribe messages = %d, want 1", count)
	}
}

func TestStaleUnsubscribeDoesNotRemoveReplacementHandler(t *testing.T) {
	name := uniqueRuntimeTestName("replacement-handler")
	first := RegisterHandler(name, func(map[string]any) {})
	second := RegisterHandler(name, func(map[string]any) {})
	first()
	mu.RLock()
	_, registered := handlers[name]
	mu.RUnlock()
	if !registered {
		t.Fatal("stale unsubscribe removed replacement handler")
	}
	second()
}

// Repeated identical messages must go through by default: two identical user
// actions within the dedup window (e.g. clicking +1 twice) are intentional.
func TestSendRepeatedMessagesNotDeduped(t *testing.T) {
	before := pendingCount()
	Send("CounterHost", map[string]any{"cmd": "increment"})
	Send("CounterHost", map[string]any{"cmd": "increment"})
	if got := pendingCount() - before; got != 2 {
		t.Fatalf("expected 2 queued messages, got %d", got)
	}
}

// Dedup is opt-in per channel: after EnableSendDedup identical payloads within
// the TTL window are dropped.
func TestSendDedupOptIn(t *testing.T) {
	name := uniqueRuntimeTestName("DedupHost")
	EnableSendDedup(name)
	before := pendingCount()
	Send(name, map[string]any{"cmd": "refresh"})
	Send(name, map[string]any{"cmd": "refresh"})
	if got := pendingCount() - before; got != 1 {
		t.Fatalf("expected 1 queued message after dedup, got %d", got)
	}
}

func TestSendMessageSerializesSequenceAndWrite(t *testing.T) {
	deliveryMu.Lock()
	savedNext := nextOutbound
	savedOutbox := outbox
	nextOutbound = 0
	outbox = map[uint64]message{}
	deliveryMu.Unlock()
	defer func() {
		deliveryMu.Lock()
		nextOutbound = savedNext
		outbox = savedOutbox
		deliveryMu.Unlock()
	}()

	firstEntered := make(chan struct{}, 1)
	firstRelease := make(chan struct{})
	secondEntered := make(chan struct{}, 1)
	firstDone := make(chan struct{})
	secondDone := make(chan struct{})
	firstWriter := func(_ context.Context, _ *hostConn, message wireMessage) error {
		firstEntered <- struct{}{}
		<-firstRelease
		if message.Sequence != 1 {
			t.Errorf("first sequence = %d, want 1", message.Sequence)
		}
		return nil
	}
	secondWriter := func(_ context.Context, _ *hostConn, message wireMessage) error {
		secondEntered <- struct{}{}
		if message.Sequence != 2 {
			t.Errorf("second sequence = %d, want 2", message.Sequence)
		}
		return nil
	}

	go func() {
		sendMessageWithWriter(nil, message{name: "first"}, firstWriter)
		close(firstDone)
	}()
	<-firstEntered
	go func() {
		sendMessageWithWriter(nil, message{name: "second"}, secondWriter)
		close(secondDone)
	}()

	select {
	case <-secondEntered:
		close(firstRelease)
		<-firstDone
		<-secondDone
		t.Fatal("second message reached the writer before the first completed")
	case <-time.After(50 * time.Millisecond):
	}
	close(firstRelease)
	<-firstDone
	<-secondDone
}

func TestPrepareInboundDeliveryResetsNewSessionState(t *testing.T) {
	sessionMu.Lock()
	savedSession := sessionID
	sessionID = "old-session"
	sessionMu.Unlock()
	deliveryMu.Lock()
	savedInbound := lastInbound
	savedToken := resumeToken
	lastInbound = 9
	resumeToken = "old-token"
	deliveryMu.Unlock()
	defer func() {
		sessionMu.Lock()
		sessionID = savedSession
		sessionMu.Unlock()
		deliveryMu.Lock()
		lastInbound = savedInbound
		resumeToken = savedToken
		deliveryMu.Unlock()
	}()

	prepareInboundDelivery("new-session", "")

	sessionMu.RLock()
	currentSession := sessionID
	sessionMu.RUnlock()
	deliveryMu.Lock()
	currentInbound := lastInbound
	currentToken := resumeToken
	deliveryMu.Unlock()
	if currentSession != "new-session" || currentInbound != 0 || currentToken != "" {
		t.Fatalf("delivery state was not reset: session=%q inbound=%d token=%q", currentSession, currentInbound, currentToken)
	}
}

func TestPrepareInboundDeliveryResetsRejectedResume(t *testing.T) {
	sessionMu.Lock()
	savedSession := sessionID
	sessionID = "current-session"
	sessionMu.Unlock()
	deliveryMu.Lock()
	savedInbound := lastInbound
	savedToken := resumeToken
	lastInbound = 9
	resumeToken = "old-token"
	deliveryMu.Unlock()
	defer func() {
		sessionMu.Lock()
		sessionID = savedSession
		sessionMu.Unlock()
		deliveryMu.Lock()
		lastInbound = savedInbound
		resumeToken = savedToken
		deliveryMu.Unlock()
	}()

	prepareInboundDelivery("current-session", "resume_rejected")

	deliveryMu.Lock()
	currentInbound := lastInbound
	currentToken := resumeToken
	deliveryMu.Unlock()
	if currentInbound != 0 || currentToken != "" {
		t.Fatalf("rejected resume state was not reset: inbound=%d token=%q", currentInbound, currentToken)
	}
}

func TestResetSessionClearsIdentityDeliveryAndPendingCalls(t *testing.T) {
	lifecycleMu.Lock()
	savedGeneration := connectionGeneration.Load()
	mu.Lock()
	savedConn := conn
	savedPending := pending
	savedBinding := bindings["reset-live"]
	_, hadBinding := bindings["reset-live"]
	savedHandler := handlers["reset-live"]
	_, hadHandler := handlers["reset-live"]
	current := newHostConn()
	closed := 0
	current.shutdown = func() error { closed++; current.fail(errConnectionClosed); return nil }
	current.generation = savedGeneration
	conn = current
	pending = []message{{name: "old-user", payload: map[string]any{"cmd": "stale"}}}
	bindings["reset-live"] = componentBinding{id: "component-live"}
	handlers["reset-live"] = func(map[string]any) {}
	mu.Unlock()

	sessionMu.Lock()
	savedSession := sessionID
	sessionID = "ssc-old"
	sessionMu.Unlock()

	deliveryMu.Lock()
	savedToken := resumeToken
	savedNext := nextOutbound
	savedInbound := lastInbound
	savedOutbox := outbox
	resumeToken = "resume-old"
	nextOutbound = 4
	lastInbound = 7
	outbox = map[uint64]message{4: {name: "old-user", sequence: 4}}
	deliveryMu.Unlock()

	callMu.Lock()
	savedCalls := pendingCalls
	reply := make(chan actionReply, 1)
	pendingCalls = map[string]chan actionReply{"call-old": reply}
	callMu.Unlock()
	lifecycleMu.Unlock()

	t.Cleanup(func() {
		lifecycleMu.Lock()
		connectionGeneration.Store(savedGeneration)
		mu.Lock()
		conn = savedConn
		pending = savedPending
		if hadBinding {
			bindings["reset-live"] = savedBinding
		} else {
			delete(bindings, "reset-live")
		}
		if hadHandler {
			handlers["reset-live"] = savedHandler
		} else {
			delete(handlers, "reset-live")
		}
		mu.Unlock()
		sessionMu.Lock()
		sessionID = savedSession
		sessionMu.Unlock()
		deliveryMu.Lock()
		resumeToken = savedToken
		nextOutbound = savedNext
		lastInbound = savedInbound
		outbox = savedOutbox
		deliveryMu.Unlock()
		callMu.Lock()
		pendingCalls = savedCalls
		callMu.Unlock()
		lifecycleMu.Unlock()
	})

	ResetSession()

	if closed != 1 {
		t.Fatalf("transport closes = %d, want 1", closed)
	}
	if got := connectionGeneration.Load(); got != savedGeneration+1 {
		t.Fatalf("generation = %d, want %d", got, savedGeneration+1)
	}
	mu.RLock()
	currentConn := conn
	queued := len(pending)
	_, bindingPreserved := bindings["reset-live"]
	_, handlerPreserved := handlers["reset-live"]
	mu.RUnlock()
	if currentConn != nil || queued != 0 {
		t.Fatalf("connection state after reset: conn=%v pending=%d", currentConn, queued)
	}
	if !bindingPreserved || !handlerPreserved {
		t.Fatal("reset removed a live binding or handler")
	}
	if got := SessionID(); got != "" {
		t.Fatalf("session ID after reset = %q", got)
	}
	deliveryMu.Lock()
	gotToken, gotNext, gotInbound, gotOutbox := resumeToken, nextOutbound, lastInbound, len(outbox)
	deliveryMu.Unlock()
	if gotToken != "" || gotNext != 0 || gotInbound != 0 || gotOutbox != 0 {
		t.Fatalf("delivery after reset: token=%q next=%d inbound=%d outbox=%d", gotToken, gotNext, gotInbound, gotOutbox)
	}
	select {
	case got := <-reply:
		if !errors.Is(got.resetErr, ErrSessionReset) {
			t.Fatalf("pending call error = %v, want ErrSessionReset", got.resetErr)
		}
	default:
		t.Fatal("reset did not fail the pending call")
	}
}

func TestOldGenerationFrameIsRejectedAfterSessionReset(t *testing.T) {
	lifecycleMu.Lock()
	savedGeneration := connectionGeneration.Load()
	connectionGeneration.Store(savedGeneration + 1)
	mu.Lock()
	savedHandler := handlers["old-generation"]
	_, hadHandler := handlers["old-generation"]
	deliveries := 0
	handlers["old-generation"] = func(map[string]any) { deliveries++ }
	mu.Unlock()
	lifecycleMu.Unlock()
	t.Cleanup(func() {
		lifecycleMu.Lock()
		connectionGeneration.Store(savedGeneration)
		mu.Lock()
		if hadHandler {
			handlers["old-generation"] = savedHandler
		} else {
			delete(handlers, "old-generation")
		}
		mu.Unlock()
		lifecycleMu.Unlock()
	})

	stale := newHostConn()
	stale.generation = savedGeneration
	stale.deliver([]byte(`{"component":"old-generation","payload":{"value":"stale"},"sequence":1}`))
	err := readLoop(context.Background(), stale)
	if !errors.Is(err, ErrSessionReset) {
		t.Fatalf("stale frame error = %v, want ErrSessionReset", err)
	}
	if deliveries != 0 {
		t.Fatalf("stale frame deliveries = %d, want 0", deliveries)
	}
}
