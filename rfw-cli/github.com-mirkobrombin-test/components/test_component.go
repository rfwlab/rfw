//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/framework"
)

//go:embed templates/test_component.rtml
var testComponentTpl []byte

type TestComponent struct {
	*framework.HTMLComponent
}

func NewTestComponent() *TestComponent {
	component := &TestComponent{
		HTMLComponent: framework.NewHTMLComponent("TestComponent", testComponentTpl, nil),
	}
	component.Init(nil)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Test Component",
	})
	component.AddDependency("header", headerComponent)

	return component
}
