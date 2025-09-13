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
	dom.RegisterHandlerFunc("fire", c.fire)
	return c
}

func (c *ToastComponent) fire() { toast.Push("Hello!") }
