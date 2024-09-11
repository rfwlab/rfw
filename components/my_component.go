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
	templateLoader := framework.NewTemplateLoader(myComponentTemplate)
	template, err := templateLoader.LoadComponentTemplate("my_component")
	if err != nil {
		panic(fmt.Sprintf("Error loading template: %v", err))
	}
	return &MyComponent{Template: template}
}

func (c *MyComponent) Render() string {
	return fmt.Sprintf(c.Template, "MyComponent render")
}
