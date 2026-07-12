// Command server is a minimal TodoMVC backend for the htmx benchmark.
// It renders HTML fragments for add/toggle/delete/filter and serves a
// polled live-update counter, mirroring how htmx apps are typically built.
package main

import (
	"flag"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type todo struct {
	ID    int
	Title string
	Done  bool
}

type app struct {
	mu     sync.Mutex
	todos  map[int]*todo
	order  []int
	nextID int
	filter string
	live   int
	dir    string
}

func (a *app) renderList(b *strings.Builder) {
	b.WriteString(`<ul id="todo-list">`)
	for _, id := range a.order {
		t := a.todos[id]
		if (a.filter == "active" && t.Done) || (a.filter == "done" && !t.Done) {
			continue
		}
		checked, class := "", ""
		if t.Done {
			checked = " checked"
			class = ` class="done"`
		}
		fmt.Fprintf(b,
			`<li%s data-id="%d"><input type="checkbox" hx-post="/toggle/%d" hx-target="#main" hx-swap="outerHTML"%s><span>%s</span><button class="destroy" hx-delete="/todos/%d" hx-target="#main" hx-swap="outerHTML">x</button></li>`,
			class, t.ID, t.ID, checked, html.EscapeString(t.Title), t.ID)
	}
	b.WriteString(`</ul>`)
}

func (a *app) renderMain() string {
	active := 0
	for _, t := range a.todos {
		if !t.Done {
			active++
		}
	}
	var b strings.Builder
	b.WriteString(`<div id="main">`)
	a.renderList(&b)
	fmt.Fprintf(&b, `<footer><span id="todo-count">%d items left</span>`, active)
	for _, f := range []string{"all", "active", "done"} {
		fmt.Fprintf(&b, `<button hx-get="/filter/%s" hx-target="#main" hx-swap="outerHTML">%s</button>`, f, strings.Title(f))
	}
	b.WriteString(`</footer></div>`)
	return b.String()
}

func (a *app) page() string {
	return `<!doctype html>
<html lang="en">
<head><meta charset="UTF-8"><title>htmx TodoMVC bench</title><script src="/htmx.min.js"></script></head>
<body>
<div class="todoapp">
<h1>todos</h1>
<form class="entry" hx-post="/todos" hx-target="#main" hx-swap="outerHTML">
<input id="new-todo" name="title" type="text" placeholder="What needs to be done?" autocomplete="off">
<button id="add-todo" type="submit">Add</button>
</form>
` + a.renderMain() + `
<div class="live">
<button id="start-live" hx-post="/live/start" hx-target="#live-zone" hx-swap="innerHTML">Start live</button>
<div id="live-zone"><span id="live-counter">0</span></div>
</div>
</div>
</body>
</html>`
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8084", "listen address")
	flag.Parse()

	dir, _ := os.Getwd()
	a := &app{todos: map[int]*todo{}, filter: "all", dir: dir}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, a.page())
	})

	mux.HandleFunc("GET /htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(a.dir, "htmx.min.js"))
	})

	mux.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		title := r.FormValue("title")
		if title != "" {
			a.nextID++
			a.todos[a.nextID] = &todo{ID: a.nextID, Title: title}
			a.order = append(a.order, a.nextID)
		}
		fmt.Fprint(w, a.renderMain())
	})

	mux.HandleFunc("POST /toggle/{id}", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		id, _ := strconv.Atoi(r.PathValue("id"))
		if t, ok := a.todos[id]; ok {
			t.Done = !t.Done
		}
		fmt.Fprint(w, a.renderMain())
	})

	mux.HandleFunc("DELETE /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		id, _ := strconv.Atoi(r.PathValue("id"))
		if _, ok := a.todos[id]; ok {
			delete(a.todos, id)
			i := sort.SearchInts(a.order, id)
			if i < len(a.order) && a.order[i] == id {
				a.order = append(a.order[:i], a.order[i+1:]...)
			}
		}
		fmt.Fprint(w, a.renderMain())
	})

	mux.HandleFunc("GET /filter/{name}", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		f := r.PathValue("name")
		if f == "all" || f == "active" || f == "done" {
			a.filter = f
		}
		fmt.Fprint(w, a.renderMain())
	})

	// Live-update scenario: the button swaps in a fragment that polls
	// /live every 100ms; the server stops the polling with HTTP 286
	// after 30 ticks (3 seconds). This is idiomatic htmx polling.
	mux.HandleFunc("POST /live/start", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		a.live = 0
		a.mu.Unlock()
		fmt.Fprint(w, `<span id="live-counter" hx-get="/live" hx-trigger="every 100ms" hx-swap="outerHTML">0</span>`)
	})

	mux.HandleFunc("GET /live", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		a.live++
		n := a.live
		a.mu.Unlock()
		if n >= 30 {
			w.WriteHeader(286) // htmx: stop polling
			fmt.Fprintf(w, `<span id="live-counter">%d</span>`, n)
			return
		}
		fmt.Fprintf(w, `<span id="live-counter" hx-get="/live" hx-trigger="every 100ms" hx-swap="outerHTML">%d</span>`, n)
	})

	fmt.Println("htmx bench server listening on", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
