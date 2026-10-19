//go:build !js

package host

import (
	"context"
	"encoding/json"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestReleasedSessionRejectsSequencedMessages(t *testing.T) {
	session := AllocateSession()
	ReleaseSession(session)
	if err := session.AcceptInbound(1); err == nil {
		t.Fatal("released session accepted a sequence")
	}
}

func TestReleasedSessionRejectsUnsequencedMessages(t *testing.T) {
	session := AllocateSession()
	ReleaseSession(session)
	if err := session.AcceptInbound(0); err == nil {
		t.Fatal("released session accepted an unsequenced message")
	}
}

func TestReleasedSessionRejectsMessagesWithNoRateLimit(t *testing.T) {
	session := AllocateSession()
	ReleaseSession(session)
	if session.AllowMessage(0) {
		t.Fatal("released session allowed a message without a rate limit")
	}
}

func TestReleasedSessionRejectsMessagesWithARateLimit(t *testing.T) {
	session := AllocateSession()
	ReleaseSession(session)
	if session.AllowMessage(10) {
		t.Fatal("released session allowed a rate-limited message")
	}
}

func TestWSReleaseSessionClosesTransportAndStopsDispatch(t *testing.T) {
	var calls atomic.Int64
	const action = "test.ws.release"
	if err := RegisterAction(action, func(context.Context, *Session, struct{}) (int64, error) {
		return calls.Add(1), nil
	}); err != nil {
		t.Fatal(err)
	}
	socket, closeSocket := openProtocolSocket(t)
	defer closeSocket()
	sendProtocolMessage(t, socket, Inbound{Action: action, ID: "before", Sequence: 1})
	first := receiveProtocolMessage(t, socket)
	session, ok := SessionByID(first.Session)
	if !ok {
		t.Fatal("session was not registered")
	}
	ReleaseSession(session)
	ReleaseSession(session)
	if _, ok := SessionByID(first.Session); ok {
		t.Error("released session remained registered")
	}
	data, err := json.Marshal(Inbound{Action: action, ID: "after", Sequence: 2})
	if err != nil {
		t.Fatal(err)
	}
	_ = websocket.Message.Send(socket, data)
	if err := socket.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	var payload []byte
	err = websocket.Message.Receive(socket, &payload)
	if err == nil {
		t.Errorf("released connection received %s", payload)
	} else if failure, ok := err.(net.Error); ok && failure.Timeout() {
		t.Errorf("released connection stayed open: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("action calls = %d", got)
	}
}

func TestReleaseSessionStopsItsWriterAndLeavesOtherSessionsUsable(t *testing.T) {
	_, server, closeSocket := openWriteTestSocket(t)
	defer closeSocket()
	peerClient, peerServer, closePeer := openWriteTestSocket(t)
	defer closePeer()
	session := AllocateResumableSession(4)
	peer := AllocateSession()
	defer ReleaseSession(peer)
	BindSessionConnection(server, session)
	BindSessionConnection(peerServer, peer)
	configureConnectionWriter(server, time.Second, 4)
	value, ok := connWriters.Load(server)
	if !ok {
		t.Fatal("writer was not configured")
	}
	writer := value.(*connectionWriter)
	ReleaseSession(session)
	select {
	case <-writer.done:
	default:
		t.Fatal("released writer remained active")
	}
	if writer.beginWrite() {
		t.Error("released writer admitted another write")
	}
	if _, ok := connWriters.Load(server); ok {
		t.Error("released writer remained registered")
	}
	if _, resumed := ResumeSession(session.ResumeToken()); resumed {
		t.Error("released session resumed")
	}
	SendSessionOutbound(peerServer, peer, Outbound{Payload: "still active"})
	if message := receiveOrderedMessage(t, peerClient); message.Payload != "still active" {
		t.Fatalf("peer response = %#v", message)
	}
}
