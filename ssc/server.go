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

	fnevents "github.com/mirkobrombin/go-foundation/v2/core/events"
	fnsafemap "github.com/mirkobrombin/go-foundation/v2/core/safemap"

	"github.com/rfwlab/rfw/v2/host"
	"golang.org/x/net/websocket"
)

type SSCEvent struct {
	Component string
	Payload   map[string]any
	Session   *host.Session
}

var (
	bus     = fnevents.New()
	connMap = fnsafemap.New[string, *fnsafemap.Map[*websocket.Conn, *host.Session]]()
)

func SubscribeSSC(fn fnevents.Handler[SSCEvent], priority ...fnevents.Priority) {
	fnevents.Subscribe[SSCEvent](bus, fn, priority...)
}

func EmitSSC(ctx context.Context, event SSCEvent) error {
	return fnevents.Emit(ctx, bus, event)
}

type SSCServer struct {
	Addr string
	Root string
	Mux  *http.ServeMux

	opts []host.MuxOption
}

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
	var sfs http.Handler
	if _, err := os.Stat(staticRoot); err == nil {
		sfs = http.FileServer(http.Dir(staticRoot))
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
		if sfs != nil {
			spath := filepath.Join(staticRoot, r.URL.Path)
			if st, err := os.Stat(spath); err == nil && !st.IsDir() {
				setWasmHeaders(w, spath, r.URL.Query().Get("v") != "")
				sfs.ServeHTTP(w, r)
				return
			}
		}
		path := filepath.Join(root, r.URL.Path)
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			setWasmHeaders(w, path, r.URL.Query().Get("v") != "")
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

func (s *SSCServer) ListenAndServe() error {
	log.Printf("SSC server starting on %s", s.Addr)
	return http.ListenAndServe(s.Addr, s.Mux)
}

func wsHandler(ws *websocket.Conn, runtime *host.WSRuntime) {
	if !runtime.AcquireConnection() {
		host.SendOutbound(ws, host.Outbound{Error: host.NewActionError("connection_limit", "connection limit reached")})
		ws.Close()
		return
	}
	defer runtime.ReleaseConnection()
	runtime.ConfigureConnection(ws)

	session, err := runtime.NewSession(ws.Request())
	if err != nil {
		host.SendOutbound(ws, host.Outbound{Error: host.NewActionError("session_rejected", "session rejected")})
		ws.Close()
		return
	}
	var subscribed []string
	subscribedSet := make(map[string]struct{})
	firstMessage := true
	defer func() {
		for _, name := range subscribed {
			if m, ok := connMap.Get(name); ok {
				m.Delete(ws)
			}
		}
		ws.Close()
		host.ForgetConnection(ws)
		host.SuspendSession(session, runtime.ResumeTTL())
	}()

	for {
		var msg host.Inbound
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			if err != io.EOF {
				log.Printf("ws receive error: %v", err)
			}
			break
		}
		if firstMessage {
			firstMessage = false
			if msg.ResumeToken != "" && msg.ResumeToken != session.ResumeToken() {
				if resumed, ok := host.ResumeSession(msg.ResumeToken); ok {
					host.ReleaseSession(session)
					session = resumed
					host.ReplaySession(ws, session, msg.Ack)
				} else {
					host.SendSessionOutbound(ws, session, host.Outbound{
						Control: "resume_rejected",
						Error:   host.NewActionError("resume_rejected", "session could not be resumed"),
					})
				}
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

		if err := fnevents.Emit(context.Background(), bus, SSCEvent{
			Component: name,
			Payload:   msg.Payload,
			Session:   session,
		}); err != nil {
			log.Printf("ssc event: %v", err)
		}
		host.SendSessionOutbound(ws, session, host.Outbound{Control: "ack"})
	}
}

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

type BroadcastOption = host.BroadcastOption

func WithSessionTarget(sessionID string) host.BroadcastOption {
	return host.WithSessionTarget(sessionID)
}

func setWasmHeaders(w http.ResponseWriter, path string, versioned bool) {
	if !strings.HasSuffix(path, ".wasm") && !strings.HasSuffix(path, ".wasm.br") {
		return
	}
	h := w.Header()
	if versioned {
		h.Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		h.Set("Cache-Control", "no-cache")
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
