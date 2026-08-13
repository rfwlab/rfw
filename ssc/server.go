// Package ssc serves applications with server-side components.
package ssc

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	fnevents "github.com/mirkobrombin/go-foundation/v2/core/events"
	fnsafemap "github.com/mirkobrombin/go-foundation/v2/core/safemap"

	"github.com/rfwlab/rfw/v2/host"
	"golang.org/x/net/websocket"
)

// SSCEvent carries a component message and its session.
type SSCEvent struct {
	Component string
	Payload   map[string]any
	Session   *host.Session
}

// Event is the concise name for SSCEvent.
type Event = SSCEvent

var (
	bus     = fnevents.New()
	connMap = fnsafemap.New[string, *fnsafemap.Map[*websocket.Conn, *host.Session]]()
)

// SubscribeSSC registers an SSC event handler.
func SubscribeSSC(fn fnevents.Handler[SSCEvent], priority ...fnevents.Priority) {
	fnevents.Subscribe[SSCEvent](bus, fn, priority...)
}

// EmitSSC emits an SSC event synchronously.
func EmitSSC(ctx context.Context, event SSCEvent) error {
	return fnevents.Emit(ctx, bus, event)
}

// SSCServer serves static assets and SSC WebSocket traffic.
type SSCServer struct {
	Addr string
	Root string
	Mux  *http.ServeMux

	opts []host.MuxOption
}

// Server is the concise name for SSCServer.
type Server = SSCServer

// NewSSCServer builds an SSC server serving files from root and the WebSocket
// endpoint at /ws. Options such as host.WithAuthFunc and
// host.WithOriginAllowlist gate the WebSocket endpoint; by default it accepts
// any origin and identity.
func NewSSCServer(addr, root string, opts ...host.MuxOption) *SSCServer {
	s := &SSCServer{Addr: addr, Root: root, opts: opts}
	s.Mux = s.buildMux()
	return s
}

func (s *SSCServer) buildMux() *http.ServeMux {
	mux := http.NewServeMux()
	runtime := host.NewWSRuntime(s.opts...)
	root := host.ResolveRoot(s.Root)
	staticRoot := filepath.Join(root, "..", "static")
	fs := http.FileServer(http.Dir(root))
	rootDir := http.Dir(root)
	var sfs http.Handler
	var staticDir http.Dir
	if _, err := os.Stat(staticRoot); err == nil {
		staticDir = http.Dir(staticRoot)
		sfs = http.FileServer(staticDir)
	}
	if sfs != nil {
		mux.Handle("/static/", http.StripPrefix("/static", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setWasmHeaders(w, r.URL.Path, r.URL.Query().Get("v") != "")
			sfs.ServeHTTP(w, r)
		})))
	}
	wsGuarded := runtime.Guard(websocket.Handler(func(ws *websocket.Conn) {
		wsHandler(ws, runtime)
	}))
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsGuarded.ServeHTTP(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("RFW_DEV_BUILD") == "1" {
			w.Header().Set("Cache-Control", "no-store")
		}
		if sfs != nil {
			if regularFile(staticDir, r.URL.Path) {
				setWasmHeaders(w, r.URL.Path, r.URL.Query().Get("v") != "")
				sfs.ServeHTTP(w, r)
				return
			}
		}
		if regularFile(rootDir, r.URL.Path) {
			setWasmHeaders(w, r.URL.Path, r.URL.Query().Get("v") != "")
			fs.ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, ".wasm") || strings.HasSuffix(r.URL.Path, ".wasm.br") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
	return mux
}

// ListenAndServe starts the SSC HTTP server.
func (s *SSCServer) ListenAndServe() error {
	log.Printf("SSC server starting on %s", s.Addr)
	server := &http.Server{
		Addr:              s.Addr,
		Handler:           s.Mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server.ListenAndServe()
}

func regularFile(root http.Dir, name string) bool {
	f, err := root.Open(name)
	if err != nil {
		return false
	}
	info, statErr := f.Stat()
	closeErr := f.Close()
	return statErr == nil && closeErr == nil && !info.IsDir()
}

func wsHandler(ws *websocket.Conn, runtime *host.WSRuntime) {
	if !runtime.AcquireConnection() {
		host.SendOutbound(ws, host.Outbound{Error: host.NewActionError("connection_limit", "connection limit reached")})
		if err := ws.Close(); err != nil {
			log.Printf("close rejected websocket: %v", err)
		}
		return
	}
	defer runtime.ReleaseConnection()
	runtime.ConfigureConnection(ws)

	var session *host.Session
	var subscribed []string
	subscribedSet := make(map[string]struct{})
	defer func() {
		for _, name := range subscribed {
			if m, ok := connMap.Get(name); ok {
				m.Delete(ws)
			}
		}
		host.SuspendSession(session, runtime.ResumeTTL())
		host.ForgetConnection(ws)
		if err := ws.Close(); err != nil {
			log.Printf("close websocket: %v", err)
		}
	}()

	for {
		var msg host.Inbound
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			if err != io.EOF {
				log.Printf("ws receive error: %v", err)
			}
			break
		}
		if session == nil {
			var resumed bool
			var err error
			session, resumed, err = runtime.OpenSession(ws.Request(), msg.ResumeToken)
			if err != nil {
				host.SendOutbound(ws, host.Outbound{Error: host.NewActionError("session_rejected", "session rejected")})
				return
			}
			host.BindSessionConnection(ws, session)
			if resumed {
				host.ReplaySession(ws, session, msg.Ack)
			} else if msg.ResumeToken != "" {
				host.SendSessionOutbound(ws, session, host.Outbound{
					Control: "resume_rejected",
					Error:   host.NewActionError("resume_rejected", "session could not be resumed"),
				})
			}
		}
		session.Acknowledge(msg.Ack)
		if err := session.AcceptInbound(msg.Sequence); err != nil {
			if errors.Is(err, host.ErrDuplicateMessage) {
				continue
			}
			host.SendSessionOutbound(ws, session, host.Outbound{
				Action: msg.Action,
				ID:     msg.ID,
				Error:  host.NewActionError("sequence_gap", "client message sequence gap"),
			})
			continue
		}
		if !session.AllowMessage(runtime.MessagesPerMinute()) {
			host.SendSessionOutbound(ws, session, host.Outbound{
				Action: msg.Action,
				ID:     msg.ID,
				Error:  host.NewActionError("rate_limited", "message rate limit exceeded"),
			})
			continue
		}
		authorizeCtx, cancelAuthorize := runtime.HandlerContext(context.Background())
		authorizeErr := runtime.Authorize(authorizeCtx, session, msg)
		cancelAuthorize()
		if authorizeErr != nil {
			host.SendSessionOutbound(ws, session, host.Outbound{
				Component: msg.Component,
				Action:    msg.Action,
				ID:        msg.ID,
				Error:     host.NewActionError("forbidden", "message forbidden"),
			})
			continue
		}
		if msg.Action != "" {
			payload, actionErr := runtime.DispatchAction(context.Background(), session, msg)
			host.SendSessionOutbound(ws, session, host.Outbound{
				Action:  msg.Action,
				ID:      msg.ID,
				Payload: payload,
				Error:   actionErr,
			})
			continue
		}
		name := msg.Component
		if name == "" {
			continue
		}
		m := connMap.GetOrSet(name, fnsafemap.New[*websocket.Conn, *host.Session]())
		m.Set(ws, session)
		if _, ok := subscribedSet[name]; !ok {
			subscribedSet[name] = struct{}{}
			subscribed = append(subscribed, name)
		}

		if hc, ok := host.Get(name); ok {
			resp := hc.HandleWithSession(session, msg.Payload)
			if resp != nil {
				switch v := resp.(type) {
				case *host.InitSnapshot:
					if v != nil {
						host.SendSessionOutbound(ws, session, host.Outbound{Component: name, ID: msg.ID, Payload: map[string]any{"initSnapshot": v}})
					}
					continue
				case host.InitSnapshot:
					host.SendSessionOutbound(ws, session, host.Outbound{Component: name, ID: msg.ID, Payload: map[string]any{"initSnapshot": v}})
					continue
				default:
					host.SendSessionOutbound(ws, session, host.Outbound{Component: name, ID: msg.ID, Payload: resp})
					continue
				}
			}
			if msg.Payload != nil && msg.Payload["init"] == true {
				host.SendSessionOutbound(ws, session, host.Outbound{Component: name, Payload: map[string]any{"session": session.ID()}})
			}
		}

		if err := fnevents.Emit(context.Background(), bus, Event{
			Component: name,
			Payload:   msg.Payload,
			Session:   session,
		}); err != nil {
			log.Printf("ssc event: %v", err)
		}
		host.SendSessionOutbound(ws, session, host.Outbound{Control: "ack"})
	}
}

// Broadcast sends a payload to connected component sessions.
func Broadcast(component string, payload any, opts ...host.BroadcastOption) {
	o := host.BroadcastOptions{Session: ""}
	for _, opt := range opts {
		opt(&o)
	}
	m, ok := connMap.Get(component)
	if !ok {
		return
	}
	m.Range(func(ws *websocket.Conn, session *host.Session) bool {
		if o.Session != "" && session.ID() != o.Session {
			return true
		}
		host.SendSessionOutbound(ws, session, host.Outbound{Component: component, Payload: payload})
		return true
	})
}

// BroadcastOption configures an SSC broadcast.
type BroadcastOption = host.BroadcastOption

// WithSessionTarget limits a broadcast to one session.
func WithSessionTarget(sessionID string) host.BroadcastOption {
	return host.WithSessionTarget(sessionID)
}

func setWasmHeaders(w http.ResponseWriter, path string, versioned bool) {
	dev := os.Getenv("RFW_DEV_BUILD") == "1"
	if dev {
		w.Header().Set("Cache-Control", "no-store")
	} else if strings.Trim(path, "/") == "rfw_config.js" {
		w.Header().Set("Cache-Control", "no-cache")
	}
	if !strings.HasSuffix(path, ".wasm") && !strings.HasSuffix(path, ".wasm.br") {
		return
	}
	h := w.Header()
	if !dev {
		if versioned {
			h.Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			h.Set("Cache-Control", "no-cache")
		}
	}
	if !strings.HasSuffix(path, ".wasm.br") {
		return
	}
	h.Set("Content-Encoding", "br")
	h.Set("Content-Type", "application/wasm")
	if vary := h.Get("Vary"); vary == "" {
		h.Set("Vary", "Accept-Encoding")
	} else if !strings.Contains(vary, "Accept-Encoding") {
		h.Set("Vary", vary+", Accept-Encoding")
	}
}
