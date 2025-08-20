//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/core"
	events "github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/router"
)

//go:embed templates/home_component.rtml
var homeTpl []byte

func NewHomeComponent() *core.HTMLComponent {
	c := core.NewComponent("HomeComponent", homeTpl, nil)
	c.SetOnMount(func(cmp *core.HTMLComponent) {
		doc := js.Document()
		add := func(sel, path string) {
			if el := doc.Call("querySelector", sel); el.Truthy() {
				ch := events.Listen("click", el)
				go func(p string) {
					for evt := range ch {
						evt.Call("preventDefault")
						router.Navigate(p)
					}
				}(path)
			}
		}
		add("a[href='/']", "/")
		add("a[href='/docs/index']", "/docs/index")
		add("a[href='/docs/getting-started']", "/docs/getting-started")
	})
	return c
}
