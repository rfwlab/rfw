//go:build js && wasm

package testkit

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/js"
)

// Harness owns a component mounted into an isolated DOM container.
type Harness struct {
	component core.Component
	container dom.Element
}

// Mount renders and mounts a component into an isolated test container.
func Mount(t TestingT, component core.Component) *Harness {
	t.Helper()
	container := dom.CreateElement("div")
	container.SetAttr("data-rfw-test", component.GetID())
	dom.Doc().Body().AppendChild(container)
	dom.UpdateDOMIn(container, component.GetID(), component.Render())
	component.Mount()
	harness := &Harness{component: component, container: container}
	t.Cleanup(harness.Unmount)
	return harness
}

// Root returns the mounted component root.
func (harness *Harness) Root() dom.Element {
	if harness == nil || harness.component == nil {
		return dom.Element{}
	}
	return harness.container.Query(`[data-component-id="` + harness.component.GetID() + `"]`)
}

// Query returns the first matching element below the harness container.
func (harness *Harness) Query(selector string) dom.Element {
	if harness == nil {
		return dom.Element{}
	}
	return harness.container.Query(selector)
}

// AssertSameNode fails when selector no longer resolves to the captured DOM
// node. It is intended for reconciliation tests where equivalent HTML is not
// enough: browser state and third-party integrations depend on node identity.
func (harness *Harness) AssertSameNode(t TestingT, selector string, captured dom.Element) {
	t.Helper()
	current := harness.Query(selector)
	if current.IsNull() || current.IsUndefined() || !current.Equal(captured.Value) {
		t.Fatalf("selector %q did not preserve DOM node identity", selector)
	}
}

// AssertActive fails unless selector resolves to the document's active element.
func (harness *Harness) AssertActive(t TestingT, selector string) {
	t.Helper()
	current := harness.Query(selector)
	if current.IsNull() || current.IsUndefined() || !js.Doc().Get("activeElement").Equal(current.Value) {
		t.Fatalf("selector %q is not the active element", selector)
	}
}

// LiveValue returns the current value property of a form control. It does not
// read the value attribute, which represents only the rendered default.
func (harness *Harness) LiveValue(selector string) string {
	return harness.Query(selector).Val()
}

// HTML returns the isolated container markup.
func (harness *Harness) HTML() string {
	if harness == nil {
		return ""
	}
	return harness.container.HTML()
}

// Text returns the isolated container text content.
func (harness *Harness) Text() string {
	if harness == nil {
		return ""
	}
	return harness.container.Text()
}

// Click dispatches a browser click on the first matching element.
func (harness *Harness) Click(t TestingT, selector string) {
	t.Helper()
	element := harness.Query(selector)
	if element.IsNull() || element.IsUndefined() {
		t.Fatalf("selector %q did not match an element", selector)
	}
	element.Call("click")
}

// Input sets a value and dispatches an input event.
func (harness *Harness) Input(t TestingT, selector, value string) {
	t.Helper()
	element := harness.Query(selector)
	if element.IsNull() || element.IsUndefined() {
		t.Fatalf("selector %q did not match an element", selector)
	}
	element.SetValue(value)
	options := js.NewDict()
	options.Set("bubbles", true)
	element.Call("dispatchEvent", js.Get("Event").New("input", options.Value))
}

// Unmount releases the component and removes its isolated container.
func (harness *Harness) Unmount() {
	if harness == nil || harness.component == nil {
		return
	}
	harness.component.Unmount()
	if !harness.container.IsNull() && !harness.container.IsUndefined() {
		harness.container.Call("remove")
	}
	harness.component = nil
	harness.container = dom.Element{}
}
