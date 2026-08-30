//go:build !js

package host

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

type singleConnListener struct {
	mu        sync.Mutex
	conn      net.Conn
	closed    chan struct{}
	closeOnce sync.Once
}

func (listener *singleConnListener) Accept() (net.Conn, error) {
	listener.mu.Lock()
	if listener.conn != nil {
		conn := listener.conn
		listener.conn = nil
		listener.mu.Unlock()
		return conn, nil
	}
	listener.mu.Unlock()
	<-listener.closed
	return nil, net.ErrClosed
}

func (listener *singleConnListener) Close() error {
	listener.closeOnce.Do(func() { close(listener.closed) })
	return nil
}

func (listener *singleConnListener) Addr() net.Addr { return pipeAddr{} }

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

func openBlockedProtocolSocket(t *testing.T, runtime *WSRuntime) (*websocket.Conn, func()) {
	t.Helper()
	serverConn, clientConn := net.Pipe()
	listener := &singleConnListener{conn: serverConn, closed: make(chan struct{})}
	server := &http.Server{
		Handler: runtime.Guard(websocket.Handler(func(ws *websocket.Conn) {
			wsHandler(ws, runtime)
		})),
		ReadHeaderTimeout: time.Second,
	}
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()

	config, err := websocket.NewConfig("ws://pipe.invalid/ws", "http://pipe.invalid")
	if err != nil {
		t.Fatalf("configure pipe websocket: %v", err)
	}
	client, err := websocket.NewClient(config, clientConn)
	if err != nil {
		_ = server.Close()
		_ = listener.Close()
		t.Fatalf("open pipe websocket: %v", err)
	}
	cleanup := func() {
		_ = client.SetDeadline(time.Now())
		_ = client.Close()
		_ = server.Close()
		_ = listener.Close()
		select {
		case err := <-serveDone:
			if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
				t.Errorf("serve pipe websocket: %v", err)
			}
		case <-time.After(time.Second):
			t.Error("pipe websocket server did not stop")
		}
	}
	return client, cleanup
}

func openRuntimeServer(t *testing.T, runtime *WSRuntime) (*httptest.Server, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("/ws", runtime.Guard(websocket.Handler(func(ws *websocket.Conn) {
		wsHandler(ws, runtime)
	})))
	server := httptest.NewServer(mux)
	return server, "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
}

func sendTestInbound(t *testing.T, ws *websocket.Conn, message Inbound) {
	t.Helper()
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal inbound: %v", err)
	}
	if err := websocket.Message.Send(ws, data); err != nil {
		t.Fatalf("send inbound: %v", err)
	}
}

func waitForConnectionCount(t *testing.T, component string, count int) map[*websocket.Conn]*Session {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		connMu.RLock()
		current := make(map[*websocket.Conn]*Session, len(connections[component]))
		for ws, session := range connections[component] {
			current[ws] = session
		}
		connMu.RUnlock()
		if len(current) == count {
			return current
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("component %q connection count did not reach %d", component, count)
	return nil
}

func waitForSessionDetached(t *testing.T, session *Session) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		session.deliveryMu.Lock()
		attached := session.attached
		session.deliveryMu.Unlock()
		if !attached {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("session did not detach after its connection closed")
}

func TestWriteDeadlineClosesAClientThatDoesNotRead(t *testing.T) {
	const componentName = "write-deadline-client"
	Register(NewHostComponent(componentName, func(map[string]any) any {
		return map[string]any{"ready": true}
	}))
	runtime := NewWSRuntime(WithSSCLimits(SSCLimits{
		WriteTimeout:      25 * time.Millisecond,
		OutboundQueueSize: 4,
	}))
	client, cleanup := openBlockedProtocolSocket(t, runtime)
	defer cleanup()

	sendTestInbound(t, client, Inbound{Component: componentName, Payload: map[string]any{"init": true}})
	targets := waitForConnectionCount(t, componentName, 1)
	var session *Session
	for _, candidate := range targets {
		session = candidate
	}
	waitForSessionDetached(t, session)

	if err := client.SetReadDeadline(time.Now().Add(time.Second)); err == nil {
		var raw []byte
		if err := websocket.Message.Receive(client, &raw); err == nil {
			t.Fatalf("slow client remained open after the write deadline: %s", raw)
		}
	}
	ReleaseSession(session)
}

func TestBroadcastDoesNotWaitForSlowClientAndOverflowResumes(t *testing.T) {
	const componentName = "bounded-broadcast-client"
	Register(NewHostComponent(componentName, func(map[string]any) any {
		return map[string]any{"ready": true}
	}))
	slowRuntime := NewWSRuntime(WithSSCLimits(SSCLimits{
		WriteTimeout:      2 * time.Second,
		OutboundQueueSize: 2,
		ResumeTTL:         time.Second,
		ReplayMessages:    16,
	}))
	fastRuntime := NewWSRuntime(WithSSCLimits(SSCLimits{
		WriteTimeout:      2 * time.Second,
		OutboundQueueSize: 16,
		ResumeTTL:         time.Second,
		ReplayMessages:    16,
	}))

	slowClient, closeSlow := openBlockedProtocolSocket(t, slowRuntime)
	defer closeSlow()
	sendTestInbound(t, slowClient, Inbound{Component: componentName, Payload: map[string]any{"init": true}})
	slowTargets := waitForConnectionCount(t, componentName, 1)
	var (
		slowServer  *websocket.Conn
		slowSession *Session
	)
	for ws, session := range slowTargets {
		slowServer = ws
		slowSession = session
	}
	token := slowSession.ResumeToken()
	sessionID := slowSession.ID()
	if token == "" {
		t.Fatal("slow session was not resumable")
	}

	deadline := time.Now().Add(time.Second)
	for {
		slowSession.deliveryMu.Lock()
		sequence := slowSession.outboundSeq
		slowSession.deliveryMu.Unlock()
		writerValue, ok := connWriters.Load(slowServer)
		if ok && sequence == 1 && len(writerValue.(*connectionWriter).queue) == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("slow client writer did not enter its blocked write")
		}
		time.Sleep(time.Millisecond)
	}

	server, wsURL := openRuntimeServer(t, fastRuntime)
	defer server.Close()
	fastClient, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("dial fast websocket: %v", err)
	}
	defer closeTestResource(t, fastClient)
	sendTestInbound(t, fastClient, Inbound{Component: componentName, Payload: map[string]any{"init": true}})
	_ = receiveOrderedMessage(t, fastClient)
	waitForConnectionCount(t, componentName, 2)

	started := time.Now()
	for tick := 1; tick <= 3; tick++ {
		Broadcast(componentName, map[string]any{"tick": tick})
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("broadcast waited for the slow client: %s", elapsed)
	}
	for tick := 1; tick <= 3; tick++ {
		message := receiveOrderedMessage(t, fastClient)
		payload, ok := message.Payload.(map[string]any)
		if !ok || payload["tick"] != float64(tick) {
			t.Fatalf("fast client broadcast %d = %#v", tick, message)
		}
	}

	if err := slowClient.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set slow client read deadline: %v", err)
	}
	resyncRequired := false
	for {
		var raw []byte
		if err := websocket.Message.Receive(slowClient, &raw); err != nil {
			break
		}
		var message Outbound
		if err := json.Unmarshal(raw, &message); err != nil {
			t.Fatalf("decode slow client message: %v", err)
		}
		if message.Control == "resync_required" && message.Error != nil && message.Error.Code == "resync_required" {
			resyncRequired = true
		}
	}
	if !resyncRequired {
		t.Fatal("queue overflow closed the slow client without resync_required")
	}
	waitForSessionDetached(t, slowSession)

	resumedClient, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("dial resumed websocket: %v", err)
	}
	sendTestInbound(t, resumedClient, Inbound{Control: "ping", ResumeToken: token})
	for sequence := uint64(1); sequence <= 4; sequence++ {
		message := receiveOrderedMessage(t, resumedClient)
		if message.Session != sessionID || message.Sequence != sequence {
			t.Fatalf("resumed message %d = %#v", sequence, message)
		}
		if sequence >= 2 {
			payload, ok := message.Payload.(map[string]any)
			if !ok || payload["tick"] != float64(sequence-1) {
				t.Fatalf("replayed broadcast %d = %#v", sequence-1, message)
			}
		}
	}
	pong := receiveOrderedMessage(t, resumedClient)
	if pong.Control != "pong" || pong.Sequence != 0 {
		t.Fatalf("resume did not continue on the new connection: %#v", pong)
	}
	closeTestResource(t, resumedClient)
	waitForSessionDetached(t, slowSession)
	ReleaseSession(slowSession)
}
