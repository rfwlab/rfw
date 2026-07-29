//go:build js && wasm

package core

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

// A component whose only reference to a store is an @if condition still has to
// react to that key: without a subscription it renders once and freezes.
func TestConditionOnStoreReactsWithoutOtherBindings(t *testing.T) {
	store := state.NewStore("condonly", state.WithModule("app"))
	store.Set("chrome", "on")
	defer state.GlobalStoreManager.UnregisterStore("app", "condonly")

	// UpdateDOM resolves an unmounted component to #app, so the page needs one
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	tpl := []byte(`<root>
@if:store:app.condonly.chrome == "on"
<span id="chrome-block">visible</span>
@endif
</root>`)
	c := NewHTMLComponent("CondOnly", tpl, nil)
	c.SetComponent(c)
	c.Init(nil)

	dom.UpdateDOM(c.GetID(), c.Render())
	c.Mount()
	defer c.Unmount()

	if !strings.Contains(dom.ComponentRoot(c.GetID()).HTML(), "chrome-block") {
		t.Fatal("condition did not render the true branch")
	}

	store.Set("chrome", "off")
	if html := dom.ComponentRoot(c.GetID()).HTML(); strings.Contains(html, "chrome-block") {
		t.Fatalf("condition did not react to the store change: %s", html)
	}

	store.Set("chrome", "on")
	if html := dom.ComponentRoot(c.GetID()).HTML(); !strings.Contains(html, "chrome-block") {
		t.Fatalf("condition did not come back: %s", html)
	}
}
