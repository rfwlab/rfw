//go:build js && wasm

package router

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

// A page that fills part of itself after mount (a card injected with SetHTML,
// the common pattern for markup built from fetched data) must keep that DOM
// when the shell around the outlet re-renders.
func TestShellRerenderKeepsPageBuiltDOM(t *testing.T) {
	Reset()
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	shellStore := state.NewStore("shellpreserve", state.WithModule("app"))
	shellStore.Set("nav", []any{map[string]any{"label": "one"}})
	defer state.GlobalStoreManager.UnregisterStore("app", "shellpreserve")

	shell := core.NewHTMLComponent("PreserveShell", []byte(`<root>
@for:it in store:app.shellpreserve.nav
<span class="nav">@prop:it.label</span>
@endfor
@include:outlet
</root>`), nil)
	shell.SetComponent(shell)
	shell.AddDependency("outlet", NewOutlet())
	shell.Init(nil)

	page := core.NewHTMLComponent("PreservePage", []byte(`<root><div data-card></div></root>`), nil)
	page.SetComponent(page)
	page.Init(nil)
	Page("/preserve-test", page)

	MountRoot(shell)
	Navigate("/preserve-test")

	// the page fills its card after mount, the way a fetch callback would
	dom.Query("[data-card]").SetHTML(`<b id="page-built">built</b>`)
	if !strings.Contains(dom.ByID("app").HTML(), "page-built") {
		t.Fatal("card was not injected")
	}

	shellStore.Set("nav", []any{map[string]any{"label": "one"}, map[string]any{"label": "two"}})

	html := dom.ByID("app").HTML()
	if !strings.Contains(html, "two") {
		t.Fatalf("shell did not re-render: %s", html)
	}
	if !strings.Contains(html, "page-built") {
		t.Fatalf("shell re-render wiped the DOM the page had built: %s", html)
	}
}
