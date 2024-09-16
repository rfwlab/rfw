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

type HTMLComponent struct {
	ID           string
	Name         string
	Template     string
	TemplateFS   []byte
	Dependencies map[string]Component
	unsubscribes []func()
	Store        *Store
	Props        map[string]interface{}
}

func NewHTMLComponent(name string, templateFs []byte, props map[string]interface{}) *HTMLComponent {
	id := generateComponentID(name, props)
	return &HTMLComponent{
		ID:           id,
		Name:         name,
		TemplateFS:   templateFs,
		Dependencies: make(map[string]Component),
		Props:        props,
	}
}

func (c *HTMLComponent) Init(store *Store) {
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

func (c *HTMLComponent) Render() string {
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
	// Handle @store:storeName.varName syntax:
	// - :w stands for writeable inputs
	// - :r stands for read-only inputs (default, not required, actually not even implemented)
	renderedTemplate = replaceStorePlaceholders(renderedTemplate, c)

	// Handle @include:componentName syntax for dependencies
	renderedTemplate = replaceIncludePlaceholders(c, renderedTemplate)

	// Handle @prop:propName syntax for props
	renderedTemplate = replacePropPlaceholders(renderedTemplate, c)

	return renderedTemplate
}

func replaceIncludePlaceholders(c *HTMLComponent, renderedTemplate string) string {
	for placeholderName, dep := range c.Dependencies {
		placeholder := fmt.Sprintf("@include:%s", placeholderName)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, dep.Render())
	}
	return renderedTemplate
}

func replaceStorePlaceholders(template string, c *HTMLComponent) string {
	storeRegex := regexp.MustCompile(`@store:(\w+)\.(\w+)(:w)?`)
	return storeRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := storeRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		storeName := parts[1]
		key := parts[2]
		isWriteable := len(parts) == 4 && parts[3] == ":w"

		store := GlobalStoreManager.GetStore(storeName)
		if store != nil {
			value := store.Get(key)
			if value == nil {
				value = ""
			}

			unsubscribe := store.OnChange(key, func(newValue interface{}) {
				UpdateDOM(c.ID, c.Render())
			})
			c.unsubscribes = append(c.unsubscribes, unsubscribe)

			if isWriteable {
				return fmt.Sprintf("@store:%s.%s:w", storeName, key)
			}
			return fmt.Sprintf("%v", value)
		}
		return match
	})
}

func replacePropPlaceholders(template string, c *HTMLComponent) string {
	propRegex := regexp.MustCompile(`@prop:(\w+)`)
	return propRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := propRegex.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		propName := parts[1]
		if value, exists := c.Props[propName]; exists {
			return fmt.Sprintf("%v", value)
		}
		return match
	})
}

func (c *HTMLComponent) AddDependency(placeholderName string, dep Component) {
	if c.Dependencies == nil {
		c.Dependencies = make(map[string]Component)
	}
	if depComp, ok := dep.(*HTMLComponent); ok {
		depComp.Init(c.Store)
	}
	c.Dependencies[placeholderName] = dep
}

func (c *HTMLComponent) Unmount() {
	for _, unsubscribe := range c.unsubscribes {
		log.Printf("Unsubscribing %s from all stores", c.Name)
		unsubscribe()
	}
	c.unsubscribes = nil

	for _, dep := range c.Dependencies {
		dep.Unmount()
	}
}

func (c *HTMLComponent) Mount() {
	for _, dep := range c.Dependencies {
		dep.Mount()
	}
}

func (c *HTMLComponent) GetName() string {
	return c.Name
}

func (c *HTMLComponent) GetID() string {
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
