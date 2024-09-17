//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/test_component.rtml
var testComponentTpl []byte

type MyComponent struct {
	*framework.HTMLComponent
}

func NewTestComponent() *MyComponent {
	component := &MyComponent{
		HTMLComponent: framework.NewHTMLComponent("MyComponent", testComponentTpl, map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{
					"name": "Mario",
					"age":  30,
				},
				map[string]interface{}{
					"name": "Luigi",
					"age":  25,
				},
			},
		}),
	}
	component.Init(nil)

	store := framework.GlobalStoreManager.GetStore("default")
	store.Set("sharedState", "Initial State")

	framework.NewStore("testing")
	framework.GlobalStoreManager.GetStore("testing").Set("testingState", "Testing Initial State")

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Another Component",
	})
	component.AddDependency("header", headerComponent)

	return component
}
