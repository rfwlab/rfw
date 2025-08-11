//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/test_component.rtml
var testComponentTpl []byte

type MyComponent struct {
	*core.HTMLComponent
}

func NewTestComponent() *MyComponent {
	c := &MyComponent{
		HTMLComponent: core.NewHTMLComponent("MyComponent", testComponentTpl, map[string]interface{}{
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
	c.Init(nil)

	store := state.GlobalStoreManager.GetStore("default")
	store.Set("sharedState", "Initial State")
	store.Set("testLoop", []interface{}{
		map[string]interface{}{
			"name": "test1",
		},
		map[string]interface{}{
			"name": "test2",
		},
	})

	state.NewStore("testing")
	state.GlobalStoreManager.GetStore("testing").Set("testingState", "Testing Initial State")

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Another Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
