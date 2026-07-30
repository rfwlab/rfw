//go:build js && wasm

package hostclient

import (
	"context"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func pendingCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(pending)
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
	EnableSendDedup("DedupHost")
	before := pendingCount()
	Send("DedupHost", map[string]any{"cmd": "refresh"})
	Send("DedupHost", map[string]any{"cmd": "refresh"})
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
	firstWriter := func(_ context.Context, _ *websocket.Conn, message wireMessage) error {
		firstEntered <- struct{}{}
		<-firstRelease
		if message.Sequence != 1 {
			t.Errorf("first sequence = %d, want 1", message.Sequence)
		}
		return nil
	}
	secondWriter := func(_ context.Context, _ *websocket.Conn, message wireMessage) error {
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
