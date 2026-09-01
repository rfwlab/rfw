package host

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mirkobrombin/go-warp/v2/streambus"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	wt "github.com/quic-go/webtransport-go"
)

const streamBusPath = "/streambus"

var (
	streamEndpoints sync.Map
	streamClientsMu sync.RWMutex
	streamClients   = make(map[string]map[*streamBusConnection]*Session)
	streamID        atomic.Uint64
	streamCertHash  atomic.Value
	streamHTTP3Port atomic.Value
)

type streamBusEndpoint struct {
	runtime *WSRuntime
	bus     *streambus.InMemory
	mu      sync.RWMutex
	server  *wt.Server
}

func newStreamBusEndpoint(runtime *WSRuntime) *streamBusEndpoint {
	maximum := DefaultSSCLimits().MaxMessageBytes
	if runtime != nil && runtime.limits.MaxMessageBytes > 0 {
		maximum = runtime.limits.MaxMessageBytes
	}
	return &streamBusEndpoint{
		runtime: runtime,
		bus: streambus.NewInMemory(streambus.Config{
			DefaultBuffer: 256, MaxBuffer: 4096, ReplayCapacity: 256,
			MaxPayloadBytes: maximum,
		}),
	}
}

func registerStreamBus(mux *http.ServeMux, runtime *WSRuntime) {
	if !streamBusEnabled() {
		return
	}
	endpoint := newStreamBusEndpoint(runtime)
	mux.Handle(streamBusPath, runtime.Guard(endpoint))
	mux.HandleFunc("/__rfw/streambus-config", func(w http.ResponseWriter, _ *http.Request) {
		hash, _ := streamCertHash.Load().(string)
		port, _ := streamHTTP3Port.Load().(string)
		if hash == "" || port == "" {
			http.Error(w, "streambus certificate is not ready", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"certificateHash": hash, "port": port})
	})
	streamEndpoints.Store(mux, endpoint)
}

func (e *streamBusEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.mu.RLock()
	server := e.server
	e.mu.RUnlock()
	if server == nil {
		http.Error(w, "streambus requires HTTP/3", http.StatusUpgradeRequired)
		return
	}
	session, err := server.Upgrade(w, r)
	if err != nil {
		return
	}
	if err := e.serveSession(r, session); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, io.EOF) {
		log.Printf("streambus: %v", err)
	}
}

func (e *streamBusEndpoint) serveSession(request *http.Request, transport *wt.Session) error {
	if !e.runtime.AcquireConnection() {
		_ = transport.CloseWithError(1, "connection limit reached")
		return errors.New("streambus: connection limit reached")
	}
	defer e.runtime.ReleaseConnection()
	stream, err := transport.AcceptStream(transport.Context())
	if err != nil {
		return err
	}
	connection, err := newStreamBusConnection(transport.Context(), request, transport, stream, e.bus, e.runtime.limits.MaxMessageBytes)
	if err != nil {
		return err
	}
	defer connection.Close()
	return streamBusProtocolLoop(connection, e.runtime)
}

type streamBusConnection struct {
	request      *http.Request
	transport    *wt.Session
	stream       *wt.Stream
	reader       *bufio.Reader
	bus          *streambus.InMemory
	topic        string
	subscription *streambus.Subscription
	maximum      int
	writeMu      sync.Mutex
	closeOnce    sync.Once
	done         chan struct{}
}

func newStreamBusConnection(ctx context.Context, request *http.Request, transport *wt.Session, stream *wt.Stream, bus *streambus.InMemory, maximum int) (*streamBusConnection, error) {
	topic := fmt.Sprintf("rfw/connection/%d", streamID.Add(1))
	subscription, err := bus.Subscribe(ctx, streambus.SubscribeOptions{Topic: topic, Buffer: 256, Overflow: streambus.Block})
	if err != nil {
		return nil, err
	}
	connection := &streamBusConnection{
		request: request, transport: transport, stream: stream, reader: bufio.NewReader(stream),
		bus: bus, topic: topic, subscription: subscription, maximum: maximum, done: make(chan struct{}),
	}
	go connection.writeLoop()
	return connection, nil
}

func (c *streamBusConnection) Receive(message *Inbound) error {
	payload, err := readFrame(c.reader, c.maximum)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, message)
}

func (c *streamBusConnection) Send(out Outbound) error {
	payload, err := json.Marshal(out)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.transport.Context(), 5*time.Second)
	defer cancel()
	_, err = c.bus.Publish(ctx, streambus.Frame{
		Topic: c.topic, Payload: payload, Reliability: streambus.Reliable,
		Priority: streambus.PriorityInteractive,
	})
	return err
}

func (c *streamBusConnection) writeLoop() {
	defer close(c.done)
	for frame := range c.subscription.Frames() {
		c.writeMu.Lock()
		err := writeFrame(c.stream, frame.Payload)
		c.writeMu.Unlock()
		if err != nil {
			_ = c.transport.CloseWithError(2, "write failed")
			return
		}
	}
}

func (c *streamBusConnection) Close() {
	c.closeOnce.Do(func() {
		_ = c.transport.CloseWithError(0, "")
		_ = c.subscription.Close()
		c.writeMu.Lock()
		_ = c.stream.Close()
		c.writeMu.Unlock()
	})
}

func streamBusProtocolLoop(connection *streamBusConnection, runtime *WSRuntime) error {
	var session *Session
	var subscribed []string
	defer func() {
		streamClientsMu.Lock()
		for _, name := range subscribed {
			if set := streamClients[name]; set != nil {
				delete(set, connection)
				if len(set) == 0 {
					delete(streamClients, name)
				}
			}
		}
		streamClientsMu.Unlock()
		SuspendSession(session, runtime.ResumeTTL())
	}()

	for {
		var message Inbound
		if err := connection.Receive(&message); err != nil {
			return err
		}
		if session == nil {
			var resumed bool
			var err error
			session, resumed, err = runtime.OpenSession(connection.request, message.ResumeToken)
			if err != nil {
				_ = connection.Send(Outbound{Error: NewActionError("session_rejected", "session rejected")})
				return err
			}
			if !bindStreamBusConnection(connection, session, true) {
				_ = connection.Send(Outbound{Error: NewActionError("session_rejected", "session already has an active connection")})
				return errors.New("streambus: session already has an active connection")
			}
			if resumed {
				replayStreamBus(connection, session, message.Ack)
			} else if message.ResumeToken != "" {
				sendStreamBusSession(connection, session, Outbound{Control: "resume_rejected", Error: NewActionError("resume_rejected", "session could not be resumed")})
			}
		}
		session.Acknowledge(message.Ack)
		if message.Control == "ping" {
			sendStreamBusSession(connection, session, Outbound{Control: "pong"})
			continue
		}
		if err := session.AcceptInbound(message.Sequence); err != nil {
			if errors.Is(err, ErrDuplicateMessage) {
				continue
			}
			sendStreamBusSession(connection, session, Outbound{ID: message.ID, Action: message.Action, Error: NewActionError("sequence_gap", "client message sequence gap")})
			continue
		}
		if !session.AllowMessage(runtime.MessagesPerMinute()) {
			sendStreamBusSession(connection, session, Outbound{ID: message.ID, Action: message.Action, Error: NewActionError("rate_limited", "message rate limit exceeded")})
			continue
		}
		authorizeCtx, cancel := runtime.HandlerContext(context.Background())
		authorizeErr := runtime.Authorize(authorizeCtx, session, message)
		cancel()
		if authorizeErr != nil {
			sendStreamBusSession(connection, session, Outbound{Component: message.Component, Action: message.Action, ID: message.ID, Error: NewActionError("forbidden", "message forbidden")})
			continue
		}
		if message.Action != "" {
			payload, actionErr := runtime.DispatchAction(context.Background(), session, message)
			sendStreamBusSession(connection, session, Outbound{Action: message.Action, ID: message.ID, Payload: payload, Error: actionErr})
			continue
		}
		if message.Component != "" && message.Payload != nil && message.Payload["unsubscribe"] == true {
			streamClientsMu.Lock()
			if set := streamClients[message.Component]; set != nil {
				delete(set, connection)
				if len(set) == 0 {
					delete(streamClients, message.Component)
				}
			}
			for index, name := range subscribed {
				if name == message.Component {
					subscribed = append(subscribed[:index], subscribed[index+1:]...)
					break
				}
			}
			streamClientsMu.Unlock()
			sendStreamBusSession(connection, session, Outbound{Component: message.Component, Control: "unsubscribed"})
			continue
		}
		if component, ok := Get(message.Component); ok {
			streamClientsMu.Lock()
			if streamClients[message.Component] == nil {
				streamClients[message.Component] = make(map[*streamBusConnection]*Session)
			}
			if _, tracked := streamClients[message.Component][connection]; !tracked {
				streamClients[message.Component][connection] = session
				subscribed = append(subscribed, message.Component)
			}
			streamClientsMu.Unlock()
			response := component.HandleWithSession(session, message.Payload)
			if response != nil {
				switch value := response.(type) {
				case *InitSnapshot:
					if value != nil {
						sendStreamBusSession(connection, session, Outbound{Component: message.Component, ID: message.ID, Payload: map[string]any{"initSnapshot": value}})
					}
				case InitSnapshot:
					sendStreamBusSession(connection, session, Outbound{Component: message.Component, ID: message.ID, Payload: map[string]any{"initSnapshot": value}})
				default:
					sendStreamBusSession(connection, session, Outbound{Component: message.Component, ID: message.ID, Payload: response})
				}
				continue
			}
			if message.Payload != nil && message.Payload["init"] == true {
				sendStreamBusSession(connection, session, Outbound{Component: message.Component, Payload: map[string]any{"session": session.ID()}})
				continue
			}
		}
		sendStreamBusSession(connection, session, Outbound{Control: "ack"})
	}
}

func bindStreamBusConnection(connection *streamBusConnection, session *Session, handoff bool) bool {
	if session == nil || connection == nil {
		return false
	}
	session.outboundMu.Lock()
	defer session.outboundMu.Unlock()
	session.deliveryMu.Lock()
	active := session.attached && !session.released
	resumePending := session.resumePending
	if active && resumePending && handoff {
		session.resumePending = false
	}
	session.deliveryMu.Unlock()
	if !active || (resumePending && !handoff) {
		return false
	}
	if session.streamConnection == connection {
		return true
	}
	if session.streamConnection != nil || session.connection != nil || (session.connectionManaged && !handoff) {
		return false
	}
	session.streamConnection = connection
	session.connection = nil
	session.connectionManaged = true
	return true
}

func sendStreamBusSession(connection *streamBusConnection, session *Session, out Outbound) {
	if session == nil {
		return
	}
	session.outboundMu.Lock()
	defer session.outboundMu.Unlock()
	if session.streamConnection != connection {
		return
	}
	_ = connection.Send(session.PrepareOutbound(out))
}

func replayStreamBus(connection *streamBusConnection, session *Session, acknowledged uint64) {
	session.outboundMu.Lock()
	defer session.outboundMu.Unlock()
	if session.streamConnection != connection {
		return
	}
	messages, err := session.ReplayAfter(acknowledged)
	if err != nil {
		_ = connection.Send(session.PrepareOutbound(Outbound{Error: NewActionError("resync_required", "message history is no longer available")}))
		return
	}
	for _, message := range messages {
		_ = connection.Send(message)
	}
}

func startStreamBusHTTP3(addr string, mux *http.ServeMux, tlsConfig *tls.Config) {
	value, ok := streamEndpoints.Load(mux)
	if !ok {
		return
	}
	endpoint := value.(*streamBusEndpoint)
	port := strings.TrimPrefix(addr, ":")
	if _, parsedPort, err := net.SplitHostPort(addr); err == nil {
		port = parsedPort
	}
	streamHTTP3Port.Store(port)
	if len(tlsConfig.Certificates) > 0 && len(tlsConfig.Certificates[0].Certificate) > 0 {
		hash := sha256.Sum256(tlsConfig.Certificates[0].Certificate[0])
		streamCertHash.Store(fmt.Sprintf("%x", hash[:]))
	}
	h3 := &http3.Server{
		Addr: addr, Handler: mux, TLSConfig: tlsConfig.Clone(), EnableDatagrams: true,
		QUICConfig: &quic.Config{EnableDatagrams: true, EnableStreamResetPartialDelivery: true},
	}
	server := &wt.Server{H3: h3}
	wt.ConfigureHTTP3Server(h3)
	endpoint.mu.Lock()
	endpoint.server = server
	endpoint.mu.Unlock()
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("streambus HTTP/3 server: %v", err)
		}
	}()
}
