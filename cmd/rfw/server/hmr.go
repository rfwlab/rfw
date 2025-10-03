package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rfwlab/rfw/cmd/rfw/utils"
)

type devMessage struct {
	Type      string `json:"type"`
	Path      string `json:"path,omitempty"`
	Component string `json:"component,omitempty"`
	Markup    string `json:"markup,omitempty"`
}

func (s *Server) handleHMR(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client := make(chan []byte, 8)

	s.hmrMu.Lock()
	s.hmrClients[client] = struct{}{}
	s.hmrMu.Unlock()

	utils.Debug("hmr client connected")

	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	defer func() {
		s.hmrMu.Lock()
		delete(s.hmrClients, client)
		s.hmrMu.Unlock()
		utils.Debug("hmr client disconnected")
	}()

	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()

	for {
		select {
		case msg, ok := <-client:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ping.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) broadcast(msg devMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	s.hmrMu.Lock()
	defer s.hmrMu.Unlock()
	if len(s.hmrClients) == 0 {
		return nil
	}
	for ch := range s.hmrClients {
		select {
		case ch <- data:
		default:
			// Drop slow clients to avoid blocking rebuilds.
			delete(s.hmrClients, ch)
			close(ch)
		}
	}
	return nil
}

func (s *Server) broadcastReload(path string) error {
	rel := path
	if abs, err := filepath.Abs(path); err == nil {
		if cwd, err := filepath.Abs("."); err == nil {
			if r, err := filepath.Rel(cwd, abs); err == nil {
				rel = filepath.ToSlash(r)
			}
		}
	}
	utils.Debug(fmt.Sprintf("broadcasting reload for %s", rel))
	return s.broadcast(devMessage{Type: "reload", Path: rel})
}

func (s *Server) broadcastTemplateUpdate(path, component, markup string) error {
	rel := path
	if abs, err := filepath.Abs(path); err == nil {
		if cwd, err := filepath.Abs("."); err == nil {
			if r, err := filepath.Rel(cwd, abs); err == nil {
				rel = filepath.ToSlash(r)
			}
		}
	}
	utils.Debug(fmt.Sprintf("streaming template update for %s (%s)", rel, component))
	return s.broadcast(devMessage{Type: "rtml", Path: rel, Component: component, Markup: markup})
}
