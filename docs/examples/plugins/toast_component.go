//go:build js && wasm

package plugins

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/plugins/toast"
)

//go:embed templates/toast_component.rtml
var toastTpl []byte

type ToastComponent struct{ *core.HTMLComponent }

func NewToastComponent() *ToastComponent {
	c := &ToastComponent{}
	c.HTMLComponent = core.NewComponent("ToastComponent", toastTpl, nil)
	dom.RegisterHandlerFunc("basic", c.basic)
	dom.RegisterHandlerFunc("custom", c.custom)
	return c
}

func (c *ToastComponent) basic() {
	toast.PushOptions("Saved!", toast.Options{
		Actions: []toast.Action{
			{Label: "Undo", Handler: func() { toast.Push("Undone") }},
		},
	})
}

func (c *ToastComponent) custom() {
	tpl := func(msg string, acts []toast.Action, close func()) dom.Element {
		doc := dom.Doc()
		el := doc.CreateElement("div")
		el.AddClass("bg-green-700")
		el.AddClass("text-white")
		el.AddClass("px-4")
		el.AddClass("py-2")
		el.AddClass("mb-2")
		el.AddClass("rounded")
		span := doc.CreateElement("span")
		span.SetText(msg)
		el.Call("appendChild", span.Value)
		btn := doc.CreateElement("button")
		btn.SetText("Got it")
		btn.AddClass("ml-4")
		btn.OnClick(func(dom.Event) { close() })
		el.Call("appendChild", btn.Value)
		return el
	}
	toast.PushOptions("Custom!", toast.Options{Template: tpl})
}
