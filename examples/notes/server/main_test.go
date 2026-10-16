//go:build !js

package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rfwlab/examples/notes/model"
)

func send(a *notesAPI, method, path, body, version string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("If-Match", version)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)
	return w
}

func TestNotesCRUDRetainsStableIDsAndRefusesStaleChanges(t *testing.T) {
	a := &notesAPI{notes: []model.Note{}}
	w := send(a, "POST", "/api/notes", `{"text":"First"}`, "")
	if w.Code != 200 || len(a.notes) != 1 || a.notes[0].Version != 1 {
		t.Fatalf("create: code=%d rows=%v", w.Code, a.notes)
	}
	w = send(a, "PUT", "/api/notes/1", `{"text":"Edited"}`, "1")
	if w.Code != 200 || a.notes[0].Text != "Edited" || a.notes[0].Version != 2 {
		t.Fatalf("edit: code=%d rows=%v", w.Code, a.notes)
	}
	w = send(a, "PUT", "/api/notes/1", `{"text":"Stale"}`, "1")
	if w.Code != 409 || a.notes[0].Text != "Edited" {
		t.Fatalf("stale edit: code=%d rows=%v", w.Code, a.notes)
	}
	if w = send(a, "DELETE", "/api/notes/1", "", "1"); w.Code != 409 || len(a.notes) != 1 {
		t.Fatalf("stale delete: code=%d rows=%v", w.Code, a.notes)
	}
	if w = send(a, "GET", "/api/notes", "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Edited") || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("read: code=%d body=%q", w.Code, w.Body.String())
	}
	if w = send(a, "DELETE", "/api/notes/1", "", "2"); w.Code != 200 || len(a.notes) != 0 {
		t.Fatalf("delete: code=%d rows=%v", w.Code, a.notes)
	}
	w = send(a, "POST", "/api/notes", `{"text":"Second"}`, "")
	var notes []model.Note
	if err := json.Unmarshal(w.Body.Bytes(), &notes); err != nil || len(notes) != 1 || notes[0].ID != 2 {
		t.Fatalf("new identity: code=%d rows=%v err=%v", w.Code, notes, err)
	}
}

func TestNotesRefusesInvalidBodiesBeforeMutation(t *testing.T) {
	a := &notesAPI{notes: []model.Note{}}
	if w := send(a, "POST", "/api/notes", `{"text":""}`, ""); w.Code != 400 {
		t.Fatalf("empty text: %d", w.Code)
	}
	if w := send(a, "POST", "/api/notes", `{"text":"ok","role":"admin"}`, ""); w.Code != 400 {
		t.Fatalf("unexpected field: %d", w.Code)
	}
	if w := send(a, "POST", "/api/notes", `{"text":"ok"}{"text":"other"}`, ""); w.Code != 400 {
		t.Fatalf("trailing data: %d", w.Code)
	}
	if w := send(a, "POST", "/api/notes", `{"text":"`+strings.Repeat("x", 5000)+`"}`, ""); w.Code != 400 {
		t.Fatalf("oversized input: %d", w.Code)
	}
	if len(a.notes) != 0 || a.next != 0 {
		t.Fatalf("invalid input mutated rows=%v next=%d", a.notes, a.next)
	}
}
