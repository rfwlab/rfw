//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/my_component.html
var myComponentTemplate []byte

type MyComponent struct {
	*framework.BaseComponent
}

func NewMyComponent() *MyComponent {
	component := &MyComponent{
		BaseComponent: framework.NewBaseComponent("MyComponent", myComponentTemplate, nil),
	}
	component.Init(nil)

	store := framework.GlobalStoreManager.GetStore("sharedStateStore")
	store.Set("sharedState", "Initial State")

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Another Component",
	})
	component.RegisterChildComponent("header", headerComponent)

	return component
}
