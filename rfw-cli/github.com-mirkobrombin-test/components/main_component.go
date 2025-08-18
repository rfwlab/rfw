//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/framework"
)

//go:embed templates/main_component.rtml
var mainComponentTpl []byte

type MainComponent struct {
	*framework.HTMLComponent
}

func NewMainComponent() *MainComponent {
	component := &MainComponent{
		HTMLComponent: framework.NewHTMLComponent("MainComponent", mainComponentTpl, nil),
	}
	component.Init(nil)

	store := framework.GlobalStoreManager.GetStore("default")
	store.Set("username", "John Doe")

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Main Component",
	})
	component.AddDependency("header", headerComponent)

	return component
}
