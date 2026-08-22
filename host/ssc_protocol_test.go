//go:build !js

package host

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func openProtocolSocket(t *testing.T, opts ...MuxOption) (*websocket.Conn, func()) {
	t.Helper()
	server := httptest.NewServer(NewMux(t.TempDir(), opts...))
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	socket, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		server.Close()
		t.Fatalf("dial websocket: %v", err)
	}
	return socket, func() {
		closeTestResource(t, socket)
		server.Close()
	}
}

func sendProtocolMessage(t *testing.T, socket *websocket.Conn, message Inbound) {
	t.Helper()
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	if err := websocket.Message.Send(socket, data); err != nil {
		t.Fatalf("send message: %v", err)
	}
}

func receiveProtocolMessage(t *testing.T, socket *websocket.Conn) Outbound {
	t.Helper()
	var data []byte
	if err := websocket.Message.Receive(socket, &data); err != nil {
		t.Fatalf("receive message: %v", err)
	}
	var message Outbound
	if err := json.Unmarshal(data, &message); err != nil {
		t.Fatalf("decode message: %v", err)
	}
	return message
}

func TestWSTypedActionRejectsUnknownFields(t *testing.T) {
	type request struct {
		Value int `json:"value"`
	}
	const action = "test.ws.strict"
	if err := RegisterAction(action, func(_ context.Context, _ *Session, request request) (request, error) {
		return request, nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	socket, closeSocket := openProtocolSocket(t)
	defer closeSocket()

	sendProtocolMessage(t, socket, Inbound{
		Action:   action,
		ID:       "strict",
		Sequence: 1,
		Payload:  map[string]any{"value": 1, "admin": true},
	})
	response := receiveProtocolMessage(t, socket)
	if response.ID != "strict" || response.Error == nil || response.Error.Code != "invalid_request" {
		t.Fatalf("unexpected strict response: %#v", response)
	}
}

func TestWSMessageAuthorizationRunsBeforeAction(t *testing.T) {
	type request struct {
		Value int `json:"value"`
	}
	const action = "test.ws.authorized"
	called := false
	if err := RegisterAction(action, func(_ context.Context, _ *Session, request request) (request, error) {
		called = true
		return request, nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	socket, closeSocket := openProtocolSocket(t, WithSSCAuthorizer(func(_ context.Context, _ *Session, message Inbound) error {
		if message.Action == action {
			return errors.New("denied")
		}
		return nil
	}))
	defer closeSocket()

	sendProtocolMessage(t, socket, Inbound{Action: action, ID: "denied", Sequence: 1})
	response := receiveProtocolMessage(t, socket)
	if response.Error == nil || response.Error.Code != "forbidden" || called {
		t.Fatalf("authorization failed closed incorrectly: response=%#v called=%v", response, called)
	}
}

func TestWSRateLimitRejectsExcessMessages(t *testing.T) {
	type request struct{}
	const action = "test.ws.rate"
	if err := RegisterAction(action, func(_ context.Context, _ *Session, _ request) (string, error) {
		return "ok", nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	socket, closeSocket := openProtocolSocket(t, WithSSCLimits(SSCLimits{MessagesPerMinute: 1}))
	defer closeSocket()

	sendProtocolMessage(t, socket, Inbound{Action: action, ID: "first", Sequence: 1})
	if response := receiveProtocolMessage(t, socket); response.Error != nil {
		t.Fatalf("first message rejected: %#v", response)
	}
	sendProtocolMessage(t, socket, Inbound{Action: action, ID: "second", Sequence: 2})
	response := receiveProtocolMessage(t, socket)
	if response.Error == nil || response.Error.Code != "rate_limited" {
		t.Fatalf("rate limit did not reject second message: %#v", response)
	}
	if response.Ack != 2 {
		t.Fatalf("rate-limited message was not acknowledged: %#v", response)
	}
}

func TestWSActionTimeout(t *testing.T) {
	type request struct{}
	const action = "test.ws.timeout"
	if err := RegisterAction(action, func(_ context.Context, _ *Session, _ request) (string, error) {
		time.Sleep(50 * time.Millisecond)
		return "late", nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	socket, closeSocket := openProtocolSocket(t, WithSSCLimits(SSCLimits{HandlerTimeout: 5 * time.Millisecond}))
	defer closeSocket()

	sendProtocolMessage(t, socket, Inbound{Action: action, ID: "timeout", Sequence: 1})
	response := receiveProtocolMessage(t, socket)
	if response.Error == nil || response.Error.Code != "action_timeout" {
		t.Fatalf("slow action did not time out: %#v", response)
	}
}

func TestWSActionPanicReturnsPublicError(t *testing.T) {
	type request struct{}
	const action = "test.ws.panic"
	if err := RegisterAction(action, func(_ context.Context, _ *Session, _ request) (string, error) {
		panic("private handler detail")
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	socket, closeSocket := openProtocolSocket(t)
	defer closeSocket()

	sendProtocolMessage(t, socket, Inbound{Action: action, ID: "panic", Sequence: 1})
	response := receiveProtocolMessage(t, socket)
	if response.Error == nil || response.Error.Code != "action_failed" || response.Error.Message != "action failed" {
		t.Fatalf("panic details crossed the protocol: %#v", response)
	}
}

func TestWSRejectsOversizedFrame(t *testing.T) {
	socket, closeSocket := openProtocolSocket(t, WithSSCLimits(SSCLimits{MaxMessageBytes: 128}))
	defer closeSocket()

	data, err := json.Marshal(Inbound{
		Component: "oversized",
		Sequence:  1,
		Payload:   map[string]any{"value": strings.Repeat("x", 512)},
	})
	if err != nil {
		t.Fatalf("marshal oversized message: %v", err)
	}
	if err := websocket.Message.Send(socket, data); err != nil {
		t.Fatalf("send oversized message: %v", err)
	}
	var response []byte
	if err := websocket.Message.Receive(socket, &response); err == nil {
		t.Fatalf("oversized frame was accepted: %s", response)
	}
}

func TestWSSessionResumeReplaysUnacknowledgedResponse(t *testing.T) {
	type request struct{}
	type response struct {
		Count int `json:"count"`
	}
	const action = "test.ws.resume"
	if err := RegisterAction(action, func(_ context.Context, session *Session, _ request) (response, error) {
		count := 0
		if stored, ok := session.ContextGet("count"); ok {
			count = stored.(int)
		}
		count++
		session.ContextSet("count", count)
		return response{Count: count}, nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}

	sessionMu.RLock()
	maxSessions := len(sessions) + 1
	sessionMu.RUnlock()
	server := httptest.NewServer(NewMux(t.TempDir(), WithSSCLimits(SSCLimits{
		ResumeTTL:      time.Second,
		ReplayMessages: 8,
		MaxSessions:    maxSessions,
	})))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	dial := func() *websocket.Conn {
		socket, err := websocket.Dial(wsURL, "", server.URL)
		if err != nil {
			t.Fatalf("dial websocket: %v", err)
		}
		return socket
	}

	firstSocket := dial()
	sendProtocolMessage(t, firstSocket, Inbound{Action: action, ID: "one", Sequence: 1})
	first := receiveProtocolMessage(t, firstSocket)
	sendProtocolMessage(t, firstSocket, Inbound{Action: action, ID: "two", Sequence: 2, Ack: first.Sequence})
	second := receiveProtocolMessage(t, firstSocket)
	token := second.ResumeToken
	sessionID := second.Session
	closeTestResource(t, firstSocket)

	var (
		secondSocket *websocket.Conn
		replayed     Outbound
		current      Outbound
	)
	deadline := time.Now().Add(time.Second)
	for {
		secondSocket = dial()
		sendProtocolMessage(t, secondSocket, Inbound{
			Action:      action,
			ID:          "three",
			Sequence:    3,
			Ack:         first.Sequence,
			ResumeToken: token,
		})
		replayed = receiveProtocolMessage(t, secondSocket)
		if replayed.Session == sessionID {
			current = receiveProtocolMessage(t, secondSocket)
			break
		}
		closeTestResource(t, secondSocket)
		if time.Now().After(deadline) {
			t.Fatalf("session did not become resumable: %#v", replayed)
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer closeTestResource(t, secondSocket)
	if replayed.Sequence != second.Sequence || replayed.ID != "two" {
		t.Fatalf("unexpected replay: %#v", replayed)
	}
	if current.Session != sessionID || current.ID != "three" {
		t.Fatalf("session did not resume: %#v", current)
	}
	payload, ok := current.Payload.(map[string]any)
	if !ok || payload["count"] != float64(3) {
		t.Fatalf("session state was not retained: %#v", current.Payload)
	}
	if session, ok := SessionByID(sessionID); ok {
		ReleaseSession(session)
	}
}

func TestWSSessionResumeExpires(t *testing.T) {
	type request struct{}
	const action = "test.ws.expiry"
	if err := RegisterAction(action, func(_ context.Context, _ *Session, _ request) (string, error) {
		return "ok", nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	server := httptest.NewServer(NewMux(t.TempDir(), WithSSCLimits(SSCLimits{ResumeTTL: 10 * time.Millisecond})))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	first, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("dial first socket: %v", err)
	}
	sendProtocolMessage(t, first, Inbound{Action: action, ID: "first", Sequence: 1})
	response := receiveProtocolMessage(t, first)
	token := response.ResumeToken
	closeTestResource(t, first)
	time.Sleep(40 * time.Millisecond)

	second, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("dial second socket: %v", err)
	}
	defer closeTestResource(t, second)
	sendProtocolMessage(t, second, Inbound{
		Action:      action,
		ID:          "second",
		Sequence:    2,
		ResumeToken: token,
	})
	rejected := receiveProtocolMessage(t, second)
	if rejected.Control != "resume_rejected" || rejected.Error == nil {
		t.Fatalf("expired session resumed: %#v", rejected)
	}
}

func TestWSControlPingIsAnsweredOutOfBand(t *testing.T) {
	socket, closeSocket := openProtocolSocket(t)
	defer closeSocket()

	sendProtocolMessage(t, socket, Inbound{Control: "ping"})
	pong := receiveProtocolMessage(t, socket)
	if pong.Control != "pong" {
		t.Fatalf("unexpected control response: %#v", pong)
	}
	if pong.Sequence != 0 || pong.Session != "" {
		t.Fatalf("pong carried delivery metadata: %#v", pong)
	}

	sendProtocolMessage(t, socket, Inbound{Component: "unregistered", Sequence: 1})
	next := receiveProtocolMessage(t, socket)
	if next.Sequence != 1 {
		t.Fatalf("pong consumed an outbound sequence: %#v", next)
	}
}
