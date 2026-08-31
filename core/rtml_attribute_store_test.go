//go:build js && wasm

package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

// A read-only @store binding is a text binding: the renderer substitutes a
// <span> for it. Inside a quoted attribute value that span closes the attribute
// early, so the browser gets a half-built <a> without its href and the rest of
// the start tag lands in the document as the literal text `yes" href=...`.
// Rendering an unsupported binding has to fail closed: report an actionable
// error and keep the broken element out of the DOM. Writable :w placeholders in
// form attributes stay covered by the golden tests.
func TestReadOnlyStoreInAttributeFailsClosed(t *testing.T) {
	store := state.NewStore("attrbind", state.WithModule("app"))
	store.Set("historyShow", "yes")
	defer state.GlobalStoreManager.UnregisterStore("app", "attrbind")

	// UpdateDOM resolves an unmounted component to #app, so the page needs one
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	var reported []string
	stopErrors := OnError(func(err any, ctx string) {
		reported = append(reported, fmt.Sprintf("%s: %v", ctx, err))
	})
	defer stopErrors()

	// The same key read as text is supported and must keep rendering.
	text := NewHTMLComponent("AttrStoreText", []byte(`<root><p>@store:app.attrbind.historyShow</p></root>`), nil)
	text.SetComponent(text)
	text.Init(nil)
	dom.UpdateDOM(text.GetID(), text.Render())
	text.Mount()
	if html := dom.ComponentRoot(text.GetID()).HTML(); !strings.Contains(html, `<span data-store="app.attrbind.historyShow">yes</span>`) {
		t.Fatalf("supported text binding did not render: %s", html)
	}
	if len(reported) != 0 {
		t.Fatalf("text binding reported errors: %v", reported)
	}
	text.Unmount()

	tpl := []byte(`<root><a data-show="@store:app.attrbind.historyShow" href="/opportunities/history">History</a><section data-show="@store:app.attrbind.historyShow" id="history-section">rows</section></root>`)
	c := NewHTMLComponent("AttrStoreAttr", tpl, nil)
	c.SetComponent(c)
	c.Init(nil)

	dom.UpdateDOM(c.GetID(), c.Render())
	c.Mount()
	defer c.Unmount()

	root := dom.ComponentRoot(c.GetID())
	html := root.HTML()
	txt := root.Text()

	if len(reported) == 0 {
		t.Fatalf("unsupported attribute binding rendered without reporting an error: %s", html)
	}
	// actionable: the report names the binding that cannot be substituted and
	// the attribute it was written into
	report := strings.Join(reported, "\n")
	if !strings.Contains(report, "app.attrbind.historyShow") || !strings.Contains(report, "data-show") {
		t.Fatalf("error report is not actionable: %s", report)
	}

	if n := root.QueryAll("a").Length(); n != 0 {
		t.Fatalf("half-built anchor reached the DOM (%d): %s", n, html)
	}
	if n := root.QueryAll("section").Length(); n != 0 {
		t.Fatalf("half-built section reached the DOM (%d): %s", n, html)
	}
	for _, leak := range []string{"data-store", "@store:"} {
		if strings.Contains(html, leak) {
			t.Fatalf("attribute substitution leaked %q into the DOM: %s", leak, html)
		}
	}
	// the tail of the start tag must not survive as document text
	for _, leak := range []string{"href=", `id="history-section"`, "yes"} {
		if strings.Contains(txt, leak) {
			t.Fatalf("corrupted literal tail %q became document text: %s", leak, txt)
		}
	}
}
