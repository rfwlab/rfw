//go:build !js

package ssc

import (
	"context"
	"net"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/host"
	"golang.org/x/net/websocket"
)

func TestSSCReleaseSessionClosesTransportAndStopsDispatch(t *testing.T) {
	var calls atomic.Int64
	const action = "test.ssc.release"
	if err := host.RegisterAction(action, func(context.Context, *host.Session, struct{}) (int64, error) {
		return calls.Add(1), nil
	}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(NewSSCServer(":0", t.TempDir()).Mux)
	defer server.Close()
	socket, err := websocket.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = socket.Close() }()
	if err := websocket.JSON.Send(socket, host.Inbound{Action: action, ID: "before", Sequence: 1}); err != nil {
		t.Fatal(err)
	}
	var first host.Outbound
	if err := websocket.JSON.Receive(socket, &first); err != nil {
		t.Fatal(err)
	}
	session, ok := host.SessionByID(first.Session)
	if !ok {
		t.Fatal("session was not registered")
	}
	host.ReleaseSession(session)
	_ = websocket.JSON.Send(socket, host.Inbound{Action: action, ID: "after", Sequence: 2})
	if err := socket.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	var response host.Outbound
	err = websocket.JSON.Receive(socket, &response)
	if err == nil {
		t.Errorf("released connection received %#v", response)
	} else if failure, ok := err.(net.Error); ok && failure.Timeout() {
		t.Errorf("released connection stayed open: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("action calls = %d", got)
	}
}
