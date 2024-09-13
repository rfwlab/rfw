//go:build js && wasm

package components

import (
	"embed"
	"fmt"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/my_component.html
var myComponentTemplate embed.FS

type MyComponent struct {
	Template string
}

func NewMyComponent() *MyComponent {
	framework.GetStore("sharedStateStore").Set("sharedState", "Initial State")

	templateLoader := framework.NewTemplateLoader(myComponentTemplate)
	template, err := templateLoader.LoadComponentTemplate("my_component")
	if err != nil {
		panic(fmt.Sprintf("Error loading template: %v", err))
	}

	return &MyComponent{
		Template: template,
	}
}

func (c *MyComponent) Render() string {
	autoReactive := framework.NewAutoReactiveComponent(c.Template)
	headerComponent := NewHeaderComponent()
	autoReactive.RegisterChildComponent("header", headerComponent)
	return autoReactive.RenderWithAutoReactive()
}
