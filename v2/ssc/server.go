package ssc

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	fncaching "github.com/mirkobrombin/go-foundation/pkg/caching"
	fnevents "github.com/mirkobrombin/go-foundation/pkg/events"
	fnsafemap "github.com/mirkobrombin/go-foundation/pkg/safemap"
	fnworker "github.com/mirkobrombin/go-foundation/pkg/worker"

	"github.com/rfwlab/rfw/v2/host"
	"golang.org/x/net/websocket"
)

type SSCEvent struct {
	Component string
	Payload   map[string]any
	Session   *host.Session
}

var (
	bus          = fnevents.New()
	connMap      = fnsafemap.New[string, *fnsafemap.Map[*websocket.Conn, *host.Session]]()
	sessionCache *fncaching.InMemoryCache[*host.Session]
	workerPool   *fnworker.Pool
)

func init() {
	sessionCache = fncaching.NewInMemory[*host.Session](
		fncaching.WithMaxEntries[*host.Session](1024),
		fncaching.WithTTL[*host.Session](10*time.Minute),
	)
	workerPool = fnworker.NewPool(4)
}

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
}

func NewSSCServer(addr, root string) *SSCServer {
	s := &SSCServer{Addr: addr, Root: root}
	s.Mux = s.buildMux()
	return s
}

func (s *SSCServer) buildMux() *http.ServeMux {
	mux := http.NewServeMux()
	root := host.ResolveRoot(s.Root)
	staticRoot := filepath.Join(root, "..", "static")
	fs := http.FileServer(http.Dir(root))
	var sfs http.Handler
	if _, err := os.Stat(staticRoot); err == nil {
		sfs = http.FileServer(http.Dir(staticRoot))
	}
	if sfs != nil {
		mux.Handle("/static/", http.StripPrefix("/static", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setWasmHeaders(w, r.URL.Path)
			sfs.ServeHTTP(w, r)
		})))
	}
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.Handler(wsHandler).ServeHTTP(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if sfs != nil {
			spath := filepath.Join(staticRoot, r.URL.Path)
			if st, err := os.Stat(spath); err == nil && !st.IsDir() {
				setWasmHeaders(w, spath)
				sfs.ServeHTTP(w, r)
				return
			}
		}
		path := filepath.Join(root, r.URL.Path)
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			setWasmHeaders(w, path)
			fs.ServeHTTP(w, r)
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

func wsHandler(ws *websocket.Conn) {
	session := host.AllocateSession()
	var subscribed []string
	defer func() {
		for _, name := range subscribed {
			if m, ok := connMap.Get(name); ok {
				m.Delete(ws)
			}
		}
		ws.Close()
		host.ReleaseSession(session)
	}()

	for {
		var msg host.Inbound
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			if err != io.EOF {
				log.Printf("ws receive error: %v", err)
			}
			break
		}
		name := msg.Component
		if name == "" {
			continue
		}
		m := connMap.GetOrSet(name, fnsafemap.New[*websocket.Conn, *host.Session]())
		m.Set(ws, session)
		subscribed = append(subscribed, name)

		if hc, ok := host.Get(name); ok {
			resp := hc.HandleWithSession(session, msg.Payload)
			if resp != nil {
				switch v := resp.(type) {
				case *host.InitSnapshot:
					if v != nil {
						sendToConn(ws, host.Outbound{Component: name, Payload: map[string]any{"initSnapshot": v}, Session: session.ID()})
					}
					continue
				case host.InitSnapshot:
					sendToConn(ws, host.Outbound{Component: name, Payload: map[string]any{"initSnapshot": v}, Session: session.ID()})
					continue
				default:
					sendToConn(ws, host.Outbound{Component: name, Payload: resp, Session: session.ID()})
					continue
				}
			}
			if msg.Payload != nil && msg.Payload["init"] == true {
				sendToConn(ws, host.Outbound{Component: name, Session: session.ID(), Payload: map[string]any{"session": session.ID()}})
			}
		}

		workerPool.Submit(func(ctx context.Context) error {
			fnevents.Emit(ctx, bus, SSCEvent{
				Component: name,
				Payload:   msg.Payload,
				Session:   session,
			})
			return nil
		})
	}
}

func sendToConn(ws *websocket.Conn, out host.Outbound) {
	data, err := json.Marshal(out)
	if err != nil {
		return
	}
	if err := websocket.Message.Send(ws, string(data)); err != nil {
		log.Printf("ssc send: %v", err)
	}
}

func Broadcast(component string, payload any, opts ...host.BroadcastOption) {
	o := host.BroadcastOptions{Session: ""}
	for _, opt := range opts {
		opt(&o)
	}
	msg := host.Outbound{Component: component, Payload: payload, Session: o.Session}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	m, ok := connMap.Get(component)
	if !ok {
		return
	}
	m.Range(func(ws *websocket.Conn, session *host.Session) bool {
		if o.Session != "" && session.ID() != o.Session {
			return true
		}
		if err := websocket.Message.Send(ws, string(data)); err != nil {
			log.Printf("broadcast send error: %v", err)
		}
		return true
	})
}

type BroadcastOption = host.BroadcastOption

func WithSessionTarget(sessionID string) host.BroadcastOption {
	return host.WithSessionTarget(sessionID)
}

func setWasmHeaders(w http.ResponseWriter, path string) {
	if !strings.HasSuffix(path, ".wasm.br") {
		return
	}
	h := w.Header()
	h.Set("Content-Encoding", "br")
	h.Set("Content-Type", "application/wasm")
	if vary := h.Get("Vary"); vary == "" {
		h.Set("Vary", "Accept-Encoding")
	} else if !strings.Contains(vary, "Accept-Encoding") {
		h.Set("Vary", vary+", Accept-Encoding")
	}
}