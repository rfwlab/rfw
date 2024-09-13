//go:build js && wasm

package components

import (
	"embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/header_component.html
var headerComponentTemplate embed.FS

type HeaderComponent struct {
	Template string
}

func NewHeaderComponent() *HeaderComponent {
	templateLoader := framework.NewTemplateLoader(headerComponentTemplate)
	template, err := templateLoader.LoadComponentTemplate("header_component")
	if err != nil {
		panic("Error loading template: " + err.Error())
	}

	return &HeaderComponent{
		Template: template,
	}
}

func (c *HeaderComponent) Render() string {
	autoReactive := framework.NewAutoReactiveComponent(c.Template)
	return autoReactive.RenderWithAutoReactive()
}
