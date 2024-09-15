//go:build js && wasm

package framework

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
)

type Component interface {
	Render() string
	Mount()
	Unmount()
	GetName() string
	GetID() string
}

type BaseComponent struct {
	ID           string
	Name         string
	Template     string
	TemplateFS   []byte
	Dependencies map[string]Component
	unsubscribes []func()
	Store        *Store
	Props        map[string]interface{}
}

func NewBaseComponent(name string, templateFs []byte, props map[string]interface{}) *BaseComponent {
	id := generateComponentID(name, props)
	return &BaseComponent{
		ID:           id,
		Name:         name,
		TemplateFS:   templateFs,
		Dependencies: make(map[string]Component),
		Props:        props,
	}
}

func (c *BaseComponent) Init(store *Store) {
	template, err := LoadComponentTemplate(c.TemplateFS)
	if err != nil {
		panic(fmt.Sprintf("Error loading template for component %s: %v", c.Name, err))
	}
	c.Template = template

	if store != nil {
		c.Store = store
	} else {
		c.Store = GlobalStoreManager.GetStore("default")
		if c.Store == nil {
			panic(fmt.Sprintf("No store provided and no default store found for component %s", c.Name))
		}
	}
}

func (c *BaseComponent) Render() string {
	for _, unsubscribe := range c.unsubscribes {
		unsubscribe()
	}
	c.unsubscribes = nil

	renderedTemplate := c.Template
	renderedTemplate = strings.Replace(renderedTemplate, "<div", fmt.Sprintf("<div data-component-id=\"%s\"", c.ID), 1)

	for key, value := range c.Props {
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, fmt.Sprintf("%v", value))
	}

	usedVariables := detectUsedVariables(renderedTemplate)
	for _, key := range usedVariables {
		store := c.Store
		value := fmt.Sprintf("%v", store.Get(key))
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, value)

		unsubscribe := store.OnChange(key, func(newValue interface{}) {
			UpdateDOM(c.ID, c.Render())
		})
		c.unsubscribes = append(c.unsubscribes, unsubscribe)
	}

	for placeholderName, dep := range c.Dependencies {
		placeholder := fmt.Sprintf("@component:%s", placeholderName)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, dep.Render())
	}

	return renderedTemplate
}

func (c *BaseComponent) AddDependency(placeholderName string, dep Component) {
	if c.Dependencies == nil {
		c.Dependencies = make(map[string]Component)
	}
	if depComp, ok := dep.(*BaseComponent); ok {
		depComp.Init(c.Store)
	}
	c.Dependencies[placeholderName] = dep
}

func (c *BaseComponent) Unmount() {
	for _, unsubscribe := range c.unsubscribes {
		log.Printf("Unsubscribing %s from all stores", c.Name)
		unsubscribe()
	}
	c.unsubscribes = nil

	for _, dep := range c.Dependencies {
		dep.Unmount()
	}
}

func (c *BaseComponent) Mount() {
	for _, dep := range c.Dependencies {
		dep.Mount()
	}
}

func (c *BaseComponent) GetName() string {
	return c.Name
}

func (c *BaseComponent) GetID() string {
	return c.ID
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

func generateComponentID(name string, props map[string]interface{}) string {
	hasher := sha1.New()
	hasher.Write([]byte(name))
	propsString := serializeProps(props)
	hasher.Write([]byte(propsString))

	return hex.EncodeToString(hasher.Sum(nil))
}

func serializeProps(props map[string]interface{}) string {
	if props == nil {
		return ""
	}

	var sb strings.Builder
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := props[k]
		sb.WriteString(fmt.Sprintf("%s=%v;", k, v))
	}

	return sb.String()
}
