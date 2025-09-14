//go:build js && wasm

package plugins

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/composition"
	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/plugins/toast"
)

//go:embed templates/toast_component.rtml
var toastTpl []byte

type ToastComponent struct{ *composition.Component }

func NewToastComponent() *ToastComponent {
	cmp := composition.Wrap(core.NewComponent("ToastComponent", toastTpl, nil))
	c := &ToastComponent{Component: cmp}
	cmp.On("basic", c.basic)
	cmp.On("custom", c.custom)
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
		div := composition.Div().Classes(
			"bg-green-700", "text-white", "px-4", "py-2", "mb-2", "rounded",
		)
		span := composition.Span().Text(msg)
		div.Element().AppendChild(span.Element())
		btn := composition.Button().Class("ml-4").Text("Got it")
		btn.Element().OnClick(func(dom.Event) { close() })
		div.Element().AppendChild(btn.Element())
		return div.Element()
	}
	toast.PushOptions("Custom!", toast.Options{Template: tpl})
}
