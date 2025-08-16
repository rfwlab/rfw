//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/stores_component.rtml
var storesComponentTpl []byte

type StoresComponent struct {
	*core.HTMLComponent
}

func NewStoresComponent() *StoresComponent {
	c := &StoresComponent{
		HTMLComponent: core.NewHTMLComponent("StoresComponent", storesComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)

	tempStore := state.NewStore("temp")
	if tempStore.Get("value") == nil {
		tempStore.Set("value", "Temporary Initial State")
	}

	permStore := state.NewStore("perm", state.WithPersistence())
	if permStore.Get("value") == nil {
		permStore.Set("value", "Persistent Initial State")
	}

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Stores",
	})
	c.AddDependency("header", headerComponent)

	return c
}
