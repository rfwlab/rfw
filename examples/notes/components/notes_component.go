//go:build js && wasm

package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"strconv"

	"github.com/rfwlab/examples/notes/model"
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/http"
)

//go:embed templates/notes_component.rtml
var notesTpl []byte

type NotesComponent struct {
	*core.HTMLComponent
	notes    []model.Note
	selected model.Note
	busy     bool
}

func NewNotesComponent() *NotesComponent {
	c := &NotesComponent{HTMLComponent: core.NewHTMLComponent("NotesComponent", notesTpl, nil)}
	c.SetComponent(c)
	dom.RegisterHandlerElem("saveNote", func(_ dom.Element, event dom.Event) {
		event.PreventDefault()
		text := c.root().Query("#note-text").Val()
		body, err := json.Marshal(struct {
			Text string `json:"text"`
		}{text})
		if err != nil {
			return
		}
		method, path := "POST", "/api/notes"
		headers := map[string]string{"Content-Type": "application/json"}
		if c.selected.ID != 0 {
			method, path = "PUT", path+"/"+strconv.Itoa(c.selected.ID)
			headers["If-Match"] = strconv.Itoa(c.selected.Version)
		}
		c.request(path, http.RequestOptions{Method: method, Headers: headers, Body: string(body)}, true)
	})
	dom.RegisterHandlerFunc("cancelNote", func() { c.reset() })
	dom.RegisterHandlerFunc("refreshNotes", func() { c.load() })
	dom.RegisterHandlerElem("editNote", func(el dom.Element, _ dom.Event) {
		id, _ := strconv.Atoi(el.Data("id"))
		for _, note := range c.notes {
			if note.ID == id {
				c.selected = note
				c.root().Query("#note-text").SetValue(note.Text)
				c.message("Editing note " + strconv.Itoa(id))
				return
			}
		}
	})
	dom.RegisterHandlerElem("deleteNote", func(el dom.Element, _ dom.Event) {
		c.request("/api/notes/"+el.Data("id"), http.RequestOptions{Method: "DELETE", Headers: map[string]string{"If-Match": el.Data("version")}}, true)
	})
	c.SetOnMount(func(*core.HTMLComponent) { c.load() })
	c.Init(nil)
	return c
}

func (c *NotesComponent) root() dom.Element   { return dom.ComponentRoot(c.ID) }
func (c *NotesComponent) message(text string) { c.root().Query("#notes-message").SetText(text) }
func (c *NotesComponent) reset() {
	if c.busy {
		return
	}
	c.selected = model.Note{}
	c.root().Query("#note-text").SetValue("")
}
func (c *NotesComponent) load() { c.request("/api/notes", http.RequestOptions{}, false) }

func (c *NotesComponent) request(path string, opts http.RequestOptions, reset bool) {
	if c.busy {
		return
	}
	c.busy = true
	c.controls(true)
	c.message("Waiting for the API...")
	// Request has no cancellation handle. Ignore a late completion after unmount,
	// using this mount's context, not a later Scope created on the same component.
	ctx := c.Scope().Context()
	http.Request(path, opts, func(status int, body string) {
		if ctx.Err() != nil {
			return
		}
		c.busy = false
		c.controls(false)
		if status == 0 {
			c.message("Connection failed. Refresh the list before retrying a write; its outcome is unknown.")
			return
		}
		if status == 409 {
			c.message("The note changed. Your draft is kept. Refresh the list and select Edit again before saving.")
			return
		}
		if status != 200 {
			c.message(fmt.Sprintf("The API refused the request (HTTP %d). Your draft is kept.", status))
			return
		}
		var notes []model.Note
		if err := json.Unmarshal([]byte(body), &notes); err != nil {
			c.message("Invalid API response. Your draft is kept.")
			return
		}
		c.notes = notes
		if reset {
			c.reset()
		}
		c.renderNotes()
		c.message("Notes loaded.")
	})
}

func (c *NotesComponent) controls(disabled bool) {
	els := c.root().QueryAll("button, input")
	for i := 0; i < els.Length(); i++ {
		el := els.Index(i)
		if disabled {
			el.SetAttr("disabled", "")
		} else {
			el.RemoveAttr("disabled")
		}
	}
}

func (c *NotesComponent) renderNotes() {
	rows := ""
	for _, note := range c.notes {
		rows += fmt.Sprintf(`<li data-note="%d"><span>%s</span> <button @on:click:editNote data-id="%d">Edit</button> <button @on:click:deleteNote data-id="%d" data-version="%d">Delete</button></li>`, note.ID, html.EscapeString(note.Text), note.ID, note.ID, note.Version)
	}
	if rows == "" {
		rows = "<li>No notes yet.</li>"
	}
	c.root().Query("#notes-list").SetHTML(dom.ExpandEvents(rows))
}
