//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"
	"strings"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/computed_component.rtml
var computedComponentTpl []byte

func NewComputedComponent() *core.HTMLComponent {
	c := core.NewComponent("ComputedComponent", computedComponentTpl, nil)

	store := state.GlobalStoreManager.GetStore("app", "default")
	if store.Get("lastChange") == nil {
		store.Set("lastChange", "")
	}
	// register computed full name
	state.Map2(store, "fullName", "first", "last", func(first, last string) string {
		return strings.TrimSpace(first + " " + last)
	})

	// show watcher side effects
	store.RegisterWatcher(state.NewWatcher([]string{"fullName"}, func(s map[string]any) {
		msg := fmt.Sprintf("name changed to %s", s["fullName"])
		fmt.Println(msg)
		store.Set("lastChange", msg)
	}))

	dom.RegisterHandlerFunc("setAda", func() {
		store.Set("first", "Ada")
		store.Set("last", "Lovelace")
	})
	dom.RegisterHandlerFunc("setGrace", func() {
		store.Set("first", "Grace")
		store.Set("last", "Hopper")
	})

	headerComponent := NewHeaderComponent(map[string]any{
		"title": "Computed Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
