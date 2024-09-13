//go:build js && wasm

package components

import (
	"embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/main_component.html
var mainComponentTemplate embed.FS

type MainComponent struct {
	Template string
}

func NewMainComponent() *MainComponent {
	templateLoader := framework.NewTemplateLoader(mainComponentTemplate)
	template, err := templateLoader.LoadComponentTemplate("main_component")
	if err != nil {
		panic("Error loading template: " + err.Error())
	}

	component := &MainComponent{
		Template: template,
	}

	return component
}

func (c *MainComponent) Render() string {
	autoReactive := framework.NewAutoReactiveComponent(c.Template)
	headerComponent := NewHeaderComponent()
	autoReactive.RegisterChildComponent("header", headerComponent)
	return autoReactive.RenderWithAutoReactive()
}
