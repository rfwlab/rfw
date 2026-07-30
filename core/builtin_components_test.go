//go:build js && wasm

package core

import (
	"strings"
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/dom"
)

func ensureAppRoot() dom.Element {
	app := dom.ByID("app")
	if app.IsNull() {
		app = dom.CreateElement("div")
		app.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(app)
	}
	app.SetHTML("")
	return app
}

func testHTMLComponent(name, template string) *HTMLComponent {
	component := NewHTMLComponent(name, []byte(template), nil)
	component.SetComponent(component)
	component.Init(nil)
	return component
}

func TestPortalMountsOutsideComponentTree(t *testing.T) {
	ensureAppRoot()
	target := dom.CreateElement("div")
	target.SetAttr("id", "portal-test-target")
	dom.Doc().Body().AppendChild(target)
	defer target.Call("remove")

	child := testHTMLComponent("PortalChild", `<root><p id="portal-content">content</p></root>`)
	portal := NewPortal("#portal-test-target", child)
	dom.UpdateDOM(portal.GetID(), portal.Render())
	portal.Mount()

	if target.Query("#portal-content").IsNull() {
		t.Fatal("portal child did not mount in target")
	}
	if !child.IsMounted() {
		t.Fatal("portal child was not mounted")
	}
	portal.Unmount()
	if !target.Query("#portal-content").IsNull() || child.IsMounted() {
		t.Fatal("portal child was not cleaned up")
	}
}

func TestKeepAlivePreservesDOMAndState(t *testing.T) {
	app := ensureAppRoot()
	child := testHTMLComponent("CachedChild", `<root><input id="cached-input" value="initial"></root>`)
	keep := NewKeepAlive(child)
	dom.UpdateDOM(keep.GetID(), keep.Render())
	keep.Mount()

	input := dom.ByID("cached-input")
	input.SetValue("edited")
	keep.Unmount()
	app.SetHTML("<p>other route</p>")

	dom.UpdateDOM(keep.GetID(), keep.Render())
	keep.Mount()
	if value := dom.ByID("cached-input").Val(); value != "edited" {
		t.Fatalf("cached DOM state was lost: %q", value)
	}
	if !child.IsMounted() {
		t.Fatal("cached child was unmounted")
	}

	keep.Dispose()
	if child.IsMounted() || !dom.ByID("cached-input").IsNull() {
		t.Fatal("disposed cache kept the child alive")
	}
}

func TestTransitionRunsEnterAndLeavePhases(t *testing.T) {
	ensureAppRoot()
	child := testHTMLComponent("TransitionChild", `<root><p>transition</p></root>`)
	transition := NewTransition(child, TransitionConfig{Duration: 20 * time.Millisecond})
	dom.UpdateDOM(transition.GetID(), transition.Render())
	transition.Mount()

	root := dom.ComponentRoot(transition.GetID())
	if !root.HasClass("rfw-enter-from") || !root.HasClass("rfw-enter-active") {
		t.Fatalf("enter phase classes missing: %s", root.Attr("class"))
	}
	transition.Unmount()
	if !root.HasClass("rfw-leave-from") || !root.HasClass("rfw-leave-active") {
		t.Fatalf("leave phase classes missing: %s", root.Attr("class"))
	}

	deadline := time.Now().Add(time.Second)
	for {
		html := dom.Doc().Body().HTML()
		if !child.IsMounted() && !strings.Contains(html, `data-component-id="`+transition.GetID()+`"`) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("transition leave did not finish: mounted=%v html=%s", child.IsMounted(), html)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestTransitionCanRemountDuringLeave(t *testing.T) {
	app := ensureAppRoot()
	child := testHTMLComponent("TransitionReturnChild", `<root><p id="transition-return">return</p></root>`)
	transition := NewTransition(child, TransitionConfig{Duration: time.Second})
	dom.UpdateDOM(transition.GetID(), transition.Render())
	transition.Mount()
	transition.Unmount()

	app.SetHTML(transition.Render())
	transition.Mount()
	if !child.IsMounted() || dom.ByID("transition-return").IsNull() {
		t.Fatal("transition did not remount during leave")
	}
	if roots := dom.QueryAll(`[data-component-id="` + transition.GetID() + `"]`); roots.Length() != 1 {
		t.Fatalf("transition left %d roots after remount", roots.Length())
	}
	transition.Dispose()
}
