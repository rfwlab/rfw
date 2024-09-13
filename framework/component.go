//go:build js && wasm

package framework

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

type Component interface {
	Render() string
	Mount()
	Unmount()
	GetName() string
}

type BaseComponent struct {
	Name         string
	Template     string
	TemplateFS   []byte
	Children     map[string]Component
	unsubscribes []func()
}

func NewBaseComponent(name string, templateFs []byte) *BaseComponent {
	return &BaseComponent{
		Name:       name,
		TemplateFS: templateFs,
		Children:   make(map[string]Component),
	}
}

func (c *BaseComponent) Init() {
	template, err := LoadComponentTemplate(c.TemplateFS)
	if err != nil {
		panic(fmt.Sprintf("Error loading template for component %s: %v", c.Name, err))
	}
	c.Template = template
	c.Children = make(map[string]Component)
}

func (c *BaseComponent) RegisterChildComponent(name string, child Component) {
	if c.Children == nil {
		c.Children = make(map[string]Component)
	}
	c.Children[name] = child
}

func (c *BaseComponent) Render() string {
	for _, unsubscribe := range c.unsubscribes {
		unsubscribe()
	}
	c.unsubscribes = nil

	usedVariables := detectUsedVariables(c.Template)
	renderedTemplate := c.Template

	for _, key := range usedVariables {
		store := GetStore("sharedStateStore")
		value := fmt.Sprintf("%v", store.Get(key))
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, value)

		unsubscribe := store.OnChange(key, func(newValue interface{}) {
			UpdateDOM(c.Render())
		})
		c.unsubscribes = append(c.unsubscribes, unsubscribe)
	}

	for name, child := range c.Children {
		childRender := child.Render()
		placeholder := fmt.Sprintf("@component:%s", name)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, childRender)
	}

	return renderedTemplate
}

func (c *BaseComponent) Unmount() {
	for _, unsubscribe := range c.unsubscribes {
		log.Printf("Unsubscribing %s from all stores", c.Name)
		unsubscribe()
	}
	c.unsubscribes = nil

	for _, child := range c.Children {
		child.Unmount()
	}
}

func (c *BaseComponent) Mount() {
	for _, child := range c.Children {
		child.Mount()
	}
}

func (c *BaseComponent) GetName() string {
	return c.Name
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
