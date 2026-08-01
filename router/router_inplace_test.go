//go:build js && wasm

package router

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
)

// Registering one component instance for a base path and its ":id" variant must
// update it in place across navigation (and browser back/forward) instead of
// unmounting and remounting an unchanged view. A remount rebuilds the
// component's DOM and wipes anything it injected after mount (e.g. a detail
// panel a fetch callback filled in), so the injected marker surviving the
// navigation is the proof it stayed mounted.
func TestSameInstanceRouteUpdatesInPlace(t *testing.T) {
	Reset()
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	shell := core.NewHTMLComponent("InplaceShell", []byte(`<root>@include:outlet</root>`), nil)
	shell.SetComponent(shell)
	shell.AddDependency("outlet", NewOutlet())
	shell.Init(nil)

	page := core.NewHTMLComponent("InplacePage", []byte(`<root><div data-inplace></div></root>`), nil)
	page.SetComponent(page)
	page.Init(nil)

	// The same instance backs both the base path and its :id variant.
	Page("/inplace", page)
	Page("/inplace/:id", page)

	MountRoot(shell)

	Navigate("/inplace")
	// The page fills itself after mount, the way a fetch callback would.
	dom.Query("[data-inplace]").SetHTML(`<b id="inplace-built">built</b>`)
	if !strings.Contains(dom.ByID("app").HTML(), "inplace-built") {
		t.Fatal("marker was not injected")
	}

	// Navigate to the :id variant of the same instance.
	Navigate("/inplace/42")
	if got := ActivePath().Get(); got != "/inplace/42" {
		t.Fatalf("active path not updated: %s", got)
	}
	if !strings.Contains(dom.ByID("app").HTML(), "inplace-built") {
		t.Fatalf("navigating to the same-instance :id route remounted the view (marker wiped): %s", dom.ByID("app").HTML())
	}

	// Back to the base path stays in place too.
	Navigate("/inplace")
	if got := ActivePath().Get(); got != "/inplace" {
		t.Fatalf("active path not restored: %s", got)
	}
	if !strings.Contains(dom.ByID("app").HTML(), "inplace-built") {
		t.Fatalf("navigating back remounted the view (marker wiped): %s", dom.ByID("app").HTML())
	}
}
