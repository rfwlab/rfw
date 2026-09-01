//go:build js && wasm

package devtools

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

func TestComponentTreeRendersFakeComponents(t *testing.T) {
	doc := dom.Doc()
	container := doc.CreateElement("div")
	doc.Query("body").AppendChild(container)
	defer container.Call("remove")

	root := core.NewHTMLComponent("RootComp", []byte("<root><div></div></root>"), nil)
	child := core.NewHTMLComponent("ChildComp", []byte("<root><span></span></root>"), nil)
	grand := core.NewHTMLComponent("GrandChildComp", []byte("<root><b></b></root>"), nil)
	child.AddDependency("grand", grand)
	root.AddDependency("child", child)

	seen := make(map[string]bool)
	buildComponentTree(container, []core.Component{root}, seen)

	text := container.Text()
	for _, want := range []string{"RootComp", "ChildComp", "GrandChildComp", root.GetID()} {
		if !strings.Contains(text, want) {
			t.Fatalf("component tree missing %q, got: %s", want, text)
		}
	}
	if !seen[grand.GetID()] {
		t.Fatalf("expected grandchild %s to be marked as seen", grand.GetID())
	}
}

func TestOverlayShowsComponentTree(t *testing.T) {
	p := New()
	core.RegisterPlugin(p)

	comp := core.NewHTMLComponent("OverlayComp", []byte("<root><div></div></root>"), nil)
	core.TriggerMount(comp)
	defer core.TriggerUnmount(comp)

	p.Show()
	defer p.panel.Call("remove")

	panel := dom.Doc().ByID("rfw-devtools")
	if !panel.Truthy() {
		t.Fatal("devtools panel not appended to document")
	}
	content := dom.Doc().ByID("rfw-devtools-content")
	if !strings.Contains(content.Text(), "OverlayComp") {
		t.Fatalf("overlay does not list mounted component, got: %s", content.Text())
	}

	p.Hide()
	if p.panel.Get("style").Get("display").String() != "none" {
		t.Fatal("panel should be hidden after Hide")
	}
}

func TestStoresTabListsStoreValues(t *testing.T) {
	doc := dom.Doc()
	container := doc.CreateElement("div")
	doc.Query("body").AppendChild(container)
	defer container.Call("remove")

	mgr := state.NewStoreManager()
	store := mgr.NewStore("session")
	store.Set("user", map[string]any{"name": "mirko"})
	store.Set("long", strings.Repeat("x", 500))

	p := New()
	p.Manager = mgr
	p.renderStores(container)

	text := container.Text()
	if !strings.Contains(text, "default/session") {
		t.Fatalf("stores tab missing store heading, got: %s", text)
	}
	if !strings.Contains(text, `user = {"name":"mirko"}`) {
		t.Fatalf("stores tab missing JSON value, got: %s", text)
	}
	if strings.Contains(text, strings.Repeat("x", valueLimit+10)) {
		t.Fatal("long value was not truncated")
	}
	if !strings.Contains(text, "…") {
		t.Fatal("truncated value missing ellipsis")
	}
}
