//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"
	"strings"

	core "github.com/rfwlab/rfw/v1/core"
	jsa "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/computed_component.rtml
var computedComponentTpl []byte

type ComputedComponent struct {
	*core.HTMLComponent
}

func NewComputedComponent() *ComputedComponent {
	c := &ComputedComponent{
		HTMLComponent: core.NewHTMLComponent("ComputedComponent", computedComponentTpl, nil),
	}
	c.Init(nil)

	store := state.GlobalStoreManager.GetStore("default")
	store.Set("lastChange", "")
	// register computed full name
	store.RegisterComputed(state.NewComputed("fullName", []string{"first", "last"}, func(s map[string]interface{}) interface{} {
		first, _ := s["first"].(string)
		last, _ := s["last"].(string)
		return strings.TrimSpace(first + " " + last)
	}))

	// show watcher side effects
	store.RegisterWatcher(state.NewWatcher([]string{"fullName"}, func(s map[string]interface{}) {
		msg := fmt.Sprintf("name changed to %s", s["fullName"])
		fmt.Println(msg)
		store.Set("lastChange", msg)
	}))

	jsa.Expose("setAda", func() {
		store.Set("first", "Ada")
		store.Set("last", "Lovelace")
	})
	jsa.Expose("setGrace", func() {
		store.Set("first", "Grace")
		store.Set("last", "Hopper")
	})

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Computed Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
