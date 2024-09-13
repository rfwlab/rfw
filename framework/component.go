//go:build js && wasm

package framework

import (
	"fmt"
	"regexp"
	"strings"
)

type AutoReactiveComponent struct {
	Template string
	Children map[string]Component // Mappa dei componenti figli
}

type Component interface {
	Render() string
}

func NewAutoReactiveComponent(template string) *AutoReactiveComponent {
	return &AutoReactiveComponent{
		Template: template,
		Children: make(map[string]Component),
	}
}

func detectUsedVariables(template string) []string {
	var re = regexp.MustCompile(`\{\{(.*?)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	var variables []string
	for _, match := range matches {
		if len(match) > 1 {
			variables = append(variables, strings.TrimSpace(match[1]))
		}
	}
	return variables
}

func (c *AutoReactiveComponent) RegisterChildComponent(name string, child Component) {
	c.Children[name] = child
}

func (c *AutoReactiveComponent) RenderWithAutoReactive() string {
	usedVariables := detectUsedVariables(c.Template)
	renderedTemplate := c.Template

	for _, key := range usedVariables {
		store := GetStore("sharedStateStore")
		value := fmt.Sprintf("%v", store.Get(key))
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, value)

		store.OnChange(key, func(newValue interface{}) {
			UpdateDOM(c.RenderWithAutoReactive())
		})
	}

	for name, child := range c.Children {
		childRender := child.Render()

		placeholder := fmt.Sprintf("@component:%s", name)
		fmt.Printf("renderedTemplate: %s\n", renderedTemplate)
		fmt.Printf("Placeholder %s is present %d times\n", placeholder, strings.Count(renderedTemplate, placeholder))
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, childRender)
	}

	return renderedTemplate
}
