//go:build js && wasm

package core

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

// The shell case: a mounted parent re-renders because one of its own store
// lists changed, and an included dependency gates its markup on another key of
// the same store. The dependency has to follow the store in the DOM, not just
// in a fresh render.
func TestMountedDependencyConditionUpdatesDOM(t *testing.T) {
	store := state.NewStore("depdom", state.WithModule("app"))
	store.Set("chrome", "on")
	store.Set("nav", []any{map[string]any{"label": "one"}})
	defer state.GlobalStoreManager.UnregisterStore("app", "depdom")

	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	child := NewHTMLComponent("DomChild", []byte(`<root>
@if:store:app.depdom.chrome == "on"
<span id="dep-block">visible</span>
@endif
</root>`), nil)
	child.SetComponent(child)
	child.Init(nil)

	parent := NewHTMLComponent("DomParent", []byte(`<root>
@for:it in store:app.depdom.nav
<span class="nav">@prop:it.label</span>
@endfor
@include:child
</root>`), nil)
	parent.SetComponent(parent)
	parent.AddDependency("child", child)
	parent.Init(nil)

	dom.UpdateDOM(parent.GetID(), parent.Render())
	parent.Mount()
	defer parent.Unmount()

	if html := dom.ComponentRoot(parent.GetID()).HTML(); !strings.Contains(html, "dep-block") {
		t.Fatalf("dependency did not render: %s", html)
	}

	store.Set("chrome", "off")
	store.Set("nav", []any{map[string]any{"label": "one"}, map[string]any{"label": "two"}})
	waitForRenderFlush()

	html := dom.ComponentRoot(parent.GetID()).HTML()
	if !strings.Contains(html, "two") {
		t.Fatalf("parent did not re-render: %s", html)
	}
	if strings.Contains(html, "dep-block") {
		t.Fatalf("dependency kept its stale markup after the store changed: %s", html)
	}
}
