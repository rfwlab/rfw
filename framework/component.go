//go:build js && wasm

package framework

import (
	"fmt"
	"strings"
)

type AutoReactiveComponent struct {
	Template string
}

func NewAutoReactiveComponent(template string) *AutoReactiveComponent {
	return &AutoReactiveComponent{
		Template: template,
	}
}

func (c *AutoReactiveComponent) RenderWithAutoReactive(stateKeys []string) string {
	store := GetGlobalStore()

	renderedTemplate := c.Template
	for _, key := range stateKeys {
		value := fmt.Sprintf("%v", store.Get(key))
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, value)

		store.OnChange(key, func(newValue interface{}) {
			UpdateDOM(c.RenderWithAutoReactive(stateKeys))
		})
	}

	return renderedTemplate
}
