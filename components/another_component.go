//go:build js && wasm

package components

import (
	"embed"
	"fmt"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/another_component.html
var anotherComponentTemplate embed.FS

type AnotherComponent struct {
	Template string
}

func NewAnotherComponent() *AnotherComponent {
	templateLoader := framework.NewTemplateLoader(anotherComponentTemplate)
	template, err := templateLoader.LoadComponentTemplate("another_component")
	if err != nil {
		panic(fmt.Sprintf("Error loading template: %v", err))
	}

	return &AnotherComponent{
		Template: template,
	}
}

func (c *AnotherComponent) Render() string {
	autoReactive := framework.NewAutoReactiveComponent(c.Template)
	return autoReactive.RenderWithAutoReactive([]string{"sharedState"})
}
