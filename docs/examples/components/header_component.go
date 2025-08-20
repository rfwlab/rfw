//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/header_component.rtml
var headerComponentTpl []byte

func NewHeaderComponent(props map[string]any) *core.HTMLComponent {
	c := core.NewComponent("HeaderComponent", headerComponentTpl, props)
	c.WithLifecycle(func(cmp *core.HTMLComponent) {
		if count, ok := cmp.Store.Get("headerMounts").(int); ok {
			cmp.Store.Set("headerMounts", count+1)
		} else {
			cmp.Store.Set("headerMounts", 1)
		}
		if _, ok := cmp.Store.Get("headerUnmounts").(int); !ok {
			cmp.Store.Set("headerUnmounts", 0)
		}
	}, func(cmp *core.HTMLComponent) {
		if count, ok := cmp.Store.Get("headerUnmounts").(int); ok {
			cmp.Store.Set("headerUnmounts", count+1)
		} else {
			cmp.Store.Set("headerUnmounts", 1)
		}
	})
	return c
}
