//go:build !js

package host

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

type blockingJSONPayload struct {
	entered chan<- struct{}
	release <-chan struct{}
	value   string
}

func (payload blockingJSONPayload) MarshalJSON() ([]byte, error) {
	payload.entered <- struct{}{}
	<-payload.release
	return json.Marshal(payload.value)
}

type signalingJSONPayload struct {
	entered chan<- struct{}
	value   string
}

func (payload signalingJSONPayload) MarshalJSON() ([]byte, error) {
	payload.entered <- struct{}{}
	return json.Marshal(payload.value)
}

func openWriteTestSocket(t *testing.T) (*websocket.Conn, *websocket.Conn, func()) {
	t.Helper()
	accepted := make(chan *websocket.Conn, 1)
	done := make(chan struct{})
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		accepted <- ws
		<-done
	}))
	client, err := websocket.Dial("ws"+strings.TrimPrefix(server.URL, "http"), "", server.URL)
	if err != nil {
		server.Close()
		t.Fatalf("dial websocket: %v", err)
	}
	serverSocket := <-accepted
	return client, serverSocket, func() {
		ForgetConnection(serverSocket)
		closeTestResource(t, client)
		closeTestResource(t, serverSocket)
		close(done)
		server.Close()
	}
}

func receiveOrderedMessage(t *testing.T, socket *websocket.Conn) Outbound {
	t.Helper()
	var raw []byte
	if err := websocket.Message.Receive(socket, &raw); err != nil {
		t.Fatalf("receive websocket message: %v", err)
	}
	var message Outbound
	if err := json.Unmarshal(raw, &message); err != nil {
		t.Fatalf("decode websocket message: %v", err)
	}
	return message
}

func TestSendSessionOutboundSerializesSequenceAndWrite(t *testing.T) {
	client, server, closeSockets := openWriteTestSocket(t)
	defer closeSockets()
	session := newSession("ordered-write")
	BindSessionConnection(server, session)
	firstEntered := make(chan struct{}, 1)
	firstRelease := make(chan struct{})
	secondEntered := make(chan struct{}, 1)
	firstDone := make(chan struct{})
	secondDone := make(chan struct{})

	go func() {
		SendSessionOutbound(server, session, Outbound{Payload: blockingJSONPayload{
			entered: firstEntered,
			release: firstRelease,
			value:   "first",
		}})
		close(firstDone)
	}()
	<-firstEntered
	go func() {
		SendSessionOutbound(server, session, Outbound{Payload: signalingJSONPayload{
			entered: secondEntered,
			value:   "second",
		}})
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

	first := receiveOrderedMessage(t, client)
	second := receiveOrderedMessage(t, client)
	if first.Sequence != 1 || second.Sequence != 2 {
		t.Fatalf("messages arrived out of order: first=%d second=%d", first.Sequence, second.Sequence)
	}
}

func TestReplaySessionDoesNotInterleaveNewMessages(t *testing.T) {
	client, server, closeSockets := openWriteTestSocket(t)
	defer closeSockets()
	session := newSession("ordered-replay", sessionOptions{replayLimit: 4})
	BindSessionConnection(server, session)
	firstEntered := make(chan struct{}, 1)
	firstRelease := make(chan struct{})
	newEntered := make(chan struct{}, 1)
	session.PrepareOutbound(Outbound{Payload: blockingJSONPayload{
		entered: firstEntered,
		release: firstRelease,
		value:   "first",
	}})
	session.PrepareOutbound(Outbound{Payload: "second"})
	replayDone := make(chan struct{})
	sendDone := make(chan struct{})

	go func() {
		ReplaySession(server, session, 0)
		close(replayDone)
	}()
	<-firstEntered
	go func() {
		SendSessionOutbound(server, session, Outbound{Payload: signalingJSONPayload{
			entered: newEntered,
			value:   "third",
		}})
		close(sendDone)
	}()

	select {
	case <-newEntered:
		close(firstRelease)
		<-replayDone
		<-sendDone
		t.Fatal("new message reached the writer during replay")
	case <-time.After(50 * time.Millisecond):
	}
	close(firstRelease)
	<-replayDone
	<-sendDone

	for sequence := uint64(1); sequence <= 3; sequence++ {
		if message := receiveOrderedMessage(t, client); message.Sequence != sequence {
			t.Fatalf("unexpected replay order: got %d want %d", message.Sequence, sequence)
		}
	}
}

func TestStaleConnectionDoesNotConsumeSequenceAfterResume(t *testing.T) {
	_, oldServer, closeOld := openWriteTestSocket(t)
	defer closeOld()
	newClient, newServer, closeNew := openWriteTestSocket(t)
	defer closeNew()
	session := AllocateResumableSession(4)
	defer ReleaseSession(session)
	BindSessionConnection(oldServer, session)
	token := session.ResumeToken()

	SuspendSession(session, time.Second)
	resumed, ok := ResumeSession(token)
	if !ok {
		t.Fatal("session did not resume")
	}
	BindSessionConnection(newServer, resumed)

	staleEntered := make(chan struct{}, 1)
	SendSessionOutbound(oldServer, resumed, Outbound{Payload: signalingJSONPayload{
		entered: staleEntered,
		value:   "stale",
	}})
	select {
	case <-staleEntered:
		t.Fatal("stale connection reached the writer")
	default:
	}

	SendSessionOutbound(newServer, resumed, Outbound{Payload: "current"})
	message := receiveOrderedMessage(t, newClient)
	if message.Sequence != 1 {
		t.Fatalf("stale connection consumed a sequence: got %d want 1", message.Sequence)
	}
}

func TestCustomHandlerDeliveryBindsAcrossResume(t *testing.T) {
	oldClient, oldServer, closeOld := openWriteTestSocket(t)
	defer closeOld()
	newClient, newServer, closeNew := openWriteTestSocket(t)
	defer closeNew()
	session := AllocateResumableSession(4)
	defer ReleaseSession(session)
	token := session.ResumeToken()

	SendSessionOutbound(oldServer, session, Outbound{Payload: "first"})
	first := receiveOrderedMessage(t, oldClient)
	if first.Sequence != 1 {
		t.Fatalf("first sequence = %d, want 1", first.Sequence)
	}
	SuspendSession(session, time.Second)
	resumed, ok := ResumeSession(token)
	if !ok {
		t.Fatal("session did not resume")
	}

	staleEntered := make(chan struct{}, 1)
	SendSessionOutbound(oldServer, resumed, Outbound{Payload: signalingJSONPayload{
		entered: staleEntered,
		value:   "stale",
	}})
	select {
	case <-staleEntered:
		t.Fatal("stale custom handler connection reached the writer")
	default:
	}

	ReplaySession(newServer, resumed, 0)
	replayed := receiveOrderedMessage(t, newClient)
	if replayed.Sequence != first.Sequence {
		t.Fatalf("replayed sequence = %d, want %d", replayed.Sequence, first.Sequence)
	}

	SendSessionOutbound(oldServer, resumed, Outbound{Payload: signalingJSONPayload{
		entered: staleEntered,
		value:   "stale",
	}})
	select {
	case <-staleEntered:
		t.Fatal("stale custom handler connection reached the writer")
	default:
	}

	SendSessionOutbound(newServer, resumed, Outbound{Payload: "second"})
	second := receiveOrderedMessage(t, newClient)
	if second.Sequence != 2 {
		t.Fatalf("second sequence = %d, want 2", second.Sequence)
	}
}

func TestManagedSessionRejectsAllPriorConnections(t *testing.T) {
	_, firstServer, closeFirst := openWriteTestSocket(t)
	defer closeFirst()
	_, secondServer, closeSecond := openWriteTestSocket(t)
	defer closeSecond()
	thirdClient, thirdServer, closeThird := openWriteTestSocket(t)
	defer closeThird()
	session := AllocateResumableSession(4)
	defer ReleaseSession(session)
	token := session.ResumeToken()
	BindSessionConnection(firstServer, session)

	SuspendSession(session, time.Second)
	resumed, ok := ResumeSession(token)
	if !ok {
		t.Fatal("first resume failed")
	}
	BindSessionConnection(secondServer, resumed)
	SuspendSession(resumed, time.Second)
	resumed, ok = ResumeSession(token)
	if !ok {
		t.Fatal("second resume failed")
	}
	BindSessionConnection(thirdServer, resumed)

	firstEntered := make(chan struct{}, 1)
	SendSessionOutbound(firstServer, resumed, Outbound{Payload: signalingJSONPayload{
		entered: firstEntered,
		value:   "first-stale",
	}})
	secondEntered := make(chan struct{}, 1)
	SendSessionOutbound(secondServer, resumed, Outbound{Payload: signalingJSONPayload{
		entered: secondEntered,
		value:   "second-stale",
	}})
	select {
	case <-firstEntered:
		t.Fatal("first stale connection reached the writer")
	case <-secondEntered:
		t.Fatal("second stale connection reached the writer")
	default:
	}

	SendSessionOutbound(thirdServer, resumed, Outbound{Payload: "current"})
	message := receiveOrderedMessage(t, thirdClient)
	if message.Sequence != 1 {
		t.Fatalf("stale connection consumed a sequence: got %d want 1", message.Sequence)
	}
}
