//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/test_component.html
var testComponentTpl []byte

type MyComponent struct {
	*framework.BaseComponent
}

func NewTestComponent() *MyComponent {
	component := &MyComponent{
		BaseComponent: framework.NewBaseComponent("MyComponent", testComponentTpl, nil),
	}
	component.Init(nil)

	store := framework.GlobalStoreManager.GetStore("default")
	store.Set("sharedState", "Initial State")

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Another Component",
	})
	component.AddDependency("header", headerComponent)

	return component
}
