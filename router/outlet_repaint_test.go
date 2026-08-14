//go:build js && wasm

package router

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

// A shell bound to a store re-renders on every write, and its fresh markup
// carries an empty outlet. The routed page has to survive that.
func TestOutletSurvivesShellRerender(t *testing.T) {
	Reset()
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	shellStore := state.NewStore("shellrepaint", state.WithModule("app"))
	shellStore.Set("title", "first")
	shellStore.Set("nav", []any{map[string]any{"label": "one"}})
	defer state.GlobalStoreManager.UnregisterStore("app", "shellrepaint")

	// a @for over a store list is the shape that re-renders the whole shell
	shell := core.NewHTMLComponent("RepaintShell", []byte(`<root>
<header>@store:app.shellrepaint.title</header>
@for:it in store:app.shellrepaint.nav
<span class="nav">@prop:it.label</span>
@endfor
@include:outlet
</root>`), nil)
	shell.SetComponent(shell)
	shell.AddDependency("outlet", NewOutlet())
	shell.Init(nil)

	page := core.NewHTMLComponent("RepaintPage", []byte(`<root><main id="routed-page">page</main></root>`), nil)
	page.SetComponent(page)
	page.Init(nil)
	Page("/repaint-test", page)

	MountRoot(shell)
	Navigate("/repaint-test")

	if html := dom.ByID("app").HTML(); !strings.Contains(html, "routed-page") {
		t.Fatalf("page did not render into the outlet: %s", html)
	}

	shellStore.Set("nav", []any{map[string]any{"label": "one"}, map[string]any{"label": "two"}})
	waitForRouterRender()

	html := dom.ByID("app").HTML()
	if !strings.Contains(html, "two") {
		t.Fatalf("shell did not re-render: %s", html)
	}
	if !strings.Contains(html, "routed-page") {
		t.Fatalf("shell re-render dropped the routed page: %s", html)
	}
}
