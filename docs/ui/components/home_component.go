//go:build js && wasm

package components

import (
	_ "embed"
	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/router"
	jst "syscall/js"
)

//go:embed templates/home_component.rtml
var homeTpl []byte

type HomeComponent struct {
	*core.HTMLComponent
	links []jst.Func
}

func NewHomeComponent() *HomeComponent {
	c := &HomeComponent{HTMLComponent: core.NewHTMLComponent("HomeComponent", homeTpl, nil)}
	c.SetComponent(c)
	c.Init(nil)
	return c
}

func (c *HomeComponent) OnMount() {
	doc := jst.Global().Get("document")
	add := func(sel, path string) {
		if el := doc.Call("querySelector", sel); el.Truthy() {
			h := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
				args[0].Call("preventDefault")
				router.Navigate(path)
				return nil
			})
			el.Call("addEventListener", "click", h)
			c.links = append(c.links, h)
		}
	}
	add("a[href='/']", "/")
	add("a[href='/docs/index']", "/docs/index")
	add("a[href='/docs/getting-started']", "/docs/getting-started")
}

func (c *HomeComponent) OnUnmount() {
	for _, h := range c.links {
		h.Release()
	}
	c.links = nil
}
