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
		BaseComponent: framework.NewBaseComponent("MyComponent", myComponentTemplate),
	}
	component.Init()

	framework.GetStore("sharedStateStore").Set("sharedState", "Initial State")

	headerComponent := NewHeaderComponent()
	component.RegisterChildComponent("header", headerComponent)

	return component
}
