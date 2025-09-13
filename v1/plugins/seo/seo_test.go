//go:build js && wasm

package seo

import (
	"testing"

	"github.com/rfwlab/rfw/v1/dom"
)

func TestSetTitleUpdatesDocument(t *testing.T) {
	p := &Plugin{}
	p.Install(nil)
	SetTitle("updated")
	if got := dom.Doc().Head().Query("title").Text(); got != "updated" {
		t.Fatalf("expected 'updated', got %q", got)
	}
}

func TestSetTitleUsesPattern(t *testing.T) {
	p := &Plugin{}
	if err := p.Build([]byte(`{"pattern":"%s - base"}`)); err != nil {
		t.Fatalf("build: %v", err)
	}
	p.Install(nil)
	SetTitle("sub")
	if got := dom.Doc().Head().Query("title").Text(); got != "sub - base" {
		t.Fatalf("expected 'sub - base', got %q", got)
	}
}

func TestSetMetaUpdatesDocument(t *testing.T) {
	p := &Plugin{}
	p.Install(nil)
	SetMeta("description", "hello")
	el := dom.Doc().Head().Query(`meta[name="description"]`)
	if got := el.Attr("content"); got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}
