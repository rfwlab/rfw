package components

import (
	"embed"
	"fmt"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/another_component.html
var anotherComponentTemplate embed.FS

type AnotherComponent struct {
	Template    string
	SharedState *framework.ReactiveVar
}

func NewAnotherComponent(sharedState *framework.ReactiveVar) *AnotherComponent {
	templateLoader := framework.NewTemplateLoader(anotherComponentTemplate)
	template, err := templateLoader.LoadComponentTemplate("another_component")
	if err != nil {
		panic(fmt.Sprintf("Error loading template: %v", err))
	}
	return &AnotherComponent{
		Template:    template,
		SharedState: sharedState,
	}
}

func (c *AnotherComponent) Render() string {
	state := c.SharedState.Get()
	return fmt.Sprintf(c.Template, state, state)
}
