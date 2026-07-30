//go:build js && wasm

package core

import (
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
)

func TestComponentDOMHookLifecycle(t *testing.T) {
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}
	component := NewHTMLComponent("Hooked", []byte("<root><p>hooked</p></root>"), nil)
	component.SetComponent(component)
	component.Init(nil)

	mounted := 0
	updated := 0
	unmounted := 0
	cleaned := 0
	component.DOMHook(dom.LifecycleHook{
		Mounted: func(root dom.Element) func() {
			if root.Attr("data-component-id") != component.ID {
				t.Fatalf("hook received wrong root: %q", root.Attr("data-component-id"))
			}
			mounted++
			return func() { cleaned++ }
		},
		Updated: func(dom.Element) {
			updated++
		},
		Unmounted: func(dom.Element) {
			unmounted++
		},
	})

	dom.UpdateDOM(component.ID, component.Render())
	component.Mount()
	dom.UpdateMountedDOM(component.ID, component.RenderFresh())
	component.Unmount()

	if mounted != 1 || updated != 1 || unmounted != 1 || cleaned != 1 {
		t.Fatalf("unexpected hook counts: mount=%d update=%d unmount=%d cleanup=%d", mounted, updated, unmounted, cleaned)
	}
}

func TestUnmountCleanupContinuesAfterLifecyclePanic(t *testing.T) {
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}
	component := NewHTMLComponent("PanicCleanup", []byte("<root></root>"), nil)
	component.SetComponent(component)
	component.Init(nil)
	cleaned := false
	component.Scope().Defer(func() { cleaned = true })
	component.SetOnUnmount(func(*HTMLComponent) { panic("unmount") })
	stopErrors := OnError(func(any, string) {})
	defer stopErrors()

	dom.UpdateDOM(component.ID, component.Render())
	component.Mount()
	component.Unmount()
	if !cleaned {
		t.Fatal("scope cleanup stopped after lifecycle panic")
	}
}
