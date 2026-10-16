//go:build !js

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rfwlab/examples/notes/model"
)

type notesAPI struct {
	mu    sync.Mutex
	notes []model.Note
	next  int
}

func (a *notesAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	a.mu.Lock()
	defer a.mu.Unlock()
	id := 0
	if r.URL.Path != "/api/notes" {
		var err error
		id, err = strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/notes/"))
		if err != nil || id < 1 {
			http.Error(w, "invalid note", 404)
			return
		}
	}
	if r.Method == "GET" && id == 0 {
		a.reply(w)
		return
	}
	if r.Method != "POST" && r.Method != "PUT" && r.Method != "DELETE" {
		w.Header().Set("Allow", "GET, POST, PUT, DELETE")
		http.Error(w, "method refused", 405)
		return
	}
	if r.Method == "POST" && id != 0 || r.Method != "POST" && id == 0 {
		http.Error(w, "invalid note", 404)
		return
	}
	idx := -1
	for i, note := range a.notes {
		if note.ID == id {
			idx = i
			break
		}
	}
	if id != 0 {
		if idx < 0 {
			http.Error(w, "note not found", 404)
			return
		}
		if r.Header.Get("If-Match") != strconv.Itoa(a.notes[idx].Version) {
			http.Error(w, "note changed", 409)
			return
		}
	}
	if r.Method == "DELETE" {
		a.notes = append(a.notes[:idx], a.notes[idx+1:]...)
		a.reply(w)
		return
	}
	var input struct {
		Text string `json:"text"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid input", 400)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "invalid input", 400)
		return
	}
	input.Text = strings.TrimSpace(input.Text)
	if len(input.Text) < 1 || len(input.Text) > 200 {
		http.Error(w, "text must contain 1..200 bytes", 400)
		return
	}
	if r.Method == "POST" {
		if len(a.notes) >= 100 {
			http.Error(w, "example limit reached", 409)
			return
		}
		a.next++
		a.notes = append(a.notes, model.Note{ID: a.next, Text: input.Text, Version: 1})
	} else {
		a.notes[idx].Text = input.Text
		a.notes[idx].Version++
	}
	a.reply(w)
}
func (a *notesAPI) reply(w http.ResponseWriter) { _ = json.NewEncoder(w).Encode(a.notes) }

func main() {
	port := flag.Int("port", 8080, "Loopback HTTP port")
	assets := flag.String("assets", "build/client", "Built RFW client directory")
	flag.Parse()
	if *port < 1 || *port > 65535 {
		log.Fatal("invalid port")
	}
	for _, name := range []string{"index.html", "app.wasm", "wasm_exec.js", "rfw_config.js", "wasm_loader.js"} {
		if _, err := os.Stat(*assets + "/" + name); err != nil {
			log.Fatal("run rfw build first: ", err)
		}
	}
	host := fmt.Sprintf("127.0.0.1:%d", *port)
	mux := http.NewServeMux()
	api := &notesAPI{notes: []model.Note{}}
	mux.Handle("/api/notes", api)
	mux.Handle("/api/notes/", api)
	files := http.FileServer(http.Dir(*assets))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/about" {
			http.ServeFile(w, r, *assets+"/index.html")
			return
		}
		files.ServeHTTP(w, r)
	})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != host || r.Header.Get("Origin") != "" && r.Header.Get("Origin") != "http://"+host {
			http.Error(w, "loopback origin required", 403)
			return
		}
		mux.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: host, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	log.Print("Example only: in-memory unauthenticated notes at http://", host)
	log.Fatal(server.ListenAndServe())
}
