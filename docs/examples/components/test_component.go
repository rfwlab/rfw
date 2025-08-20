//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/test_component.rtml
var testComponentTpl []byte

func NewTestComponent() *core.HTMLComponent {
	c := core.NewComponent("MyComponent", testComponentTpl, map[string]any{
		"items": []any{
			map[string]any{
				"name": "Mario",
				"age":  30,
			},
			map[string]any{
				"name": "Luigi",
				"age":  25,
			},
		},
		"obj": map[string]any{
			"first":  "Mario",
			"second": "Luigi",
		},
		"n": 3,
	})

	store := state.GlobalStoreManager.GetStore("app", "default")
	if store.Get("testLoop") == nil {
		store.Set("testLoop", []any{
			map[string]any{
				"name": "test1",
			},
			map[string]any{
				"name": "test2",
			},
		})
	}
	return c
}
