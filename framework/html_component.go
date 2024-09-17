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
	"syscall/js"
)

type HTMLComponent struct {
	ID                string
	Name              string
	Template          string
	TemplateFS        []byte
	Dependencies      map[string]Component
	unsubscribes      []func()
	Store             *Store
	Props             map[string]interface{}
	conditionContents map[string]ConditionContent
}

type ConditionContent struct {
	conditionStr string
	ifContent    string
	elseContent  string
}

type ConditionalBlock struct {
	Condition   string
	IfContent   string
	ElseContent string
}

type ConditionDependency struct {
	storeName string
	key       string
}

func NewHTMLComponent(name string, templateFs []byte, props map[string]interface{}) *HTMLComponent {
	id := generateComponentID(name, props)
	return &HTMLComponent{
		ID:                id,
		Name:              name,
		TemplateFS:        templateFs,
		Dependencies:      make(map[string]Component),
		Props:             props,
		conditionContents: make(map[string]ConditionContent),
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
	renderedTemplate = strings.Replace(renderedTemplate, "<root", fmt.Sprintf("<root data-component-id=\"%s\"", c.ID), 1)

	for key, value := range c.Props {
		placeholder := fmt.Sprintf("{{%s}}", key)
		renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, fmt.Sprintf("%v", value))
	}

	// Handle @include:componentName syntax for dependencies
	renderedTemplate = replaceIncludePlaceholders(c, renderedTemplate)

	// Handle @foreach:collection as item syntax
	renderedTemplate = replaceForeachPlaceholders(renderedTemplate, c)

	// Handle @store:storeName.varName syntax:
	// - :w stands for writeable inputs
	// - :r stands for read-only inputs (default, not required, actually not even implemented)
	renderedTemplate = replaceStorePlaceholders(renderedTemplate, c)

	// Handle @prop:propName syntax for props
	renderedTemplate = replacePropPlaceholders(renderedTemplate, c)

	// Handle @if:condition syntax for conditional rendering
	renderedTemplate = replaceConditionals(renderedTemplate, c)

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
				UpdateStoreBindings(c, storeName, key, newValue)
			})
			c.unsubscribes = append(c.unsubscribes, unsubscribe)

			if isWriteable {
				return match
			} else {
				return fmt.Sprintf(`<span data-store="%s.%s">%v</span>`, storeName, key, value)
			}
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

func replaceConditionals(template string, c *HTMLComponent) string {
	ifRegex := regexp.MustCompile(`(@if:.+)([\S\s]+)(@else)([\S\s]+)(@endif)`)
	template = ifRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := ifRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		conditionStr := strings.TrimSpace(parts[1])
		ifContent := parts[2]
		elseContent := parts[4]

		conditionID := fmt.Sprintf("cond-%x", sha1.Sum([]byte(conditionStr)))

		result, dependencies := evaluateCondition(conditionStr, c)

		c.conditionContents[conditionID] = ConditionContent{
			conditionStr: conditionStr,
			ifContent:    ifContent,
			elseContent:  elseContent,
		}

		for _, dep := range dependencies {
			storeName, key := dep.storeName, dep.key
			store := GlobalStoreManager.GetStore(storeName)
			if store != nil {
				unsubscribe := store.OnChange(key, func(newValue interface{}) {
					UpdateConditionBindings(c, conditionID, conditionStr)
				})
				c.unsubscribes = append(c.unsubscribes, unsubscribe)
				fmt.Printf("Registered listener for %s.%s\n", storeName, key)
			} else {
				fmt.Printf("Store %s not found\n", storeName)
			}
		}

		var content string
		if result {
			content = ifContent
		} else {
			content = elseContent
		}

		return fmt.Sprintf(`<div data-condition="%s">%s</div>`, conditionID, content)
	})

	// Manage @if without @else
	ifRegexNoElse := regexp.MustCompile(`(@if:.+)([\S\s]+)(@endif)`)
	template = ifRegexNoElse.ReplaceAllStringFunc(template, func(match string) string {
		parts := ifRegexNoElse.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		conditionStr := strings.TrimSpace(parts[1])
		ifContent := parts[2]

		conditionID := fmt.Sprintf("cond-%x", sha1.Sum([]byte(conditionStr)))

		result, dependencies := evaluateCondition(conditionStr, c)

		c.conditionContents[conditionID] = ConditionContent{
			conditionStr: conditionStr,
			ifContent:    ifContent,
			elseContent:  "",
		}

		for _, dep := range dependencies {
			storeName, key := dep.storeName, dep.key
			fmt.Printf("Registering listener for condition '%s' on store '%s', key '%s'\n", conditionStr, storeName, key)
			store := GlobalStoreManager.GetStore(storeName)
			if store != nil {
				unsubscribe := store.OnChange(key, func(newValue interface{}) {
					fmt.Printf("Listener triggered for store '%s', key '%s', new value: '%v'\n", storeName, key, newValue)
					UpdateConditionBindings(c, conditionID, conditionStr)
				})
				c.unsubscribes = append(c.unsubscribes, unsubscribe)
			} else {
				fmt.Printf("Store '%s' not found when registering listener.\n", storeName)
			}
		}

		var content string
		if result {
			content = ifContent
		} else {
			content = ""
		}

		return fmt.Sprintf(`<div data-condition="%s">%s</div>`, conditionID, content)
	})

	return template
}

func evaluateCondition(condition string, c *HTMLComponent) (bool, []ConditionDependency) {
	fmt.Printf("Evaluating condition: '%s'\n", condition)
	conditionParts := strings.Split(condition, "==")
	if len(conditionParts) != 2 {
		fmt.Println("Condition format is invalid. Expected '=='.")
		return false, nil
	}

	leftSide := strings.TrimSpace(conditionParts[0])
	leftSide = strings.Replace(leftSide, "@if:", "", 1)
	expectedValue := strings.ReplaceAll(conditionParts[1], `"`, "")
	expectedValue = strings.TrimSpace(expectedValue)

	fmt.Printf("Left side: '%s', Expected value: '%s'\n", leftSide, expectedValue)

	dependencies := []ConditionDependency{}

	if strings.HasPrefix(leftSide, "store:") {
		storeParts := strings.Split(strings.TrimPrefix(leftSide, "store:"), ".")
		if len(storeParts) == 2 {
			storeName, key := storeParts[0], storeParts[1]
			fmt.Printf("Dependency detected: Store '%s', Key '%s'\n", storeName, key)
			store := GlobalStoreManager.GetStore(storeName)
			if store != nil {
				dependencies = append(dependencies, ConditionDependency{storeName, key})
				actualValue := fmt.Sprintf("%v", store.Get(key))
				fmt.Printf("Actual value from store: '%s'\n", actualValue)
				return actualValue == expectedValue, dependencies
			} else {
				fmt.Printf("Store '%s' not found.\n", storeName)
			}
		} else {
			fmt.Println("Store parts length is not 2.")
		}
	}

	if strings.HasPrefix(leftSide, "prop:") {
		propName := strings.TrimPrefix(leftSide, "prop:")
		if value, exists := c.Props[propName]; exists {
			actualValue := fmt.Sprintf("%v", value)
			fmt.Printf("Actual value from props: '%s'\n", actualValue)
			return actualValue == expectedValue, dependencies
		}
	}

	fmt.Println("No dependencies detected.")
	return false, dependencies
}

func UpdateConditionBindings(c *HTMLComponent, conditionID, conditionStr string) {
	document := js.Global().Get("document")
	var element js.Value
	if c.ID == "" {
		element = document.Call("getElementById", "app")
	} else {
		element = document.Call("querySelector", fmt.Sprintf("[data-component-id='%s']", c.ID))
	}
	if element.IsNull() || element.IsUndefined() {
		return
	}

	selector := fmt.Sprintf(`[data-condition="%s"]`, conditionID)
	node := element.Call("querySelector", selector)
	if node.IsNull() || node.IsUndefined() {
		return
	}

	result, _ := evaluateCondition(conditionStr, c)

	conditionContent := c.conditionContents[conditionID]
	var newContent string
	if result {
		newContent = conditionContent.ifContent
	} else {
		newContent = conditionContent.elseContent
	}

	node.Set("innerHTML", newContent)

	bindStoreInputs(node)
}

func UpdateConditionsForStoreVariable(c *HTMLComponent, storeName, key string) {
	for conditionID, content := range c.conditionContents {
		dependencies, _ := getConditionDependencies(content.conditionStr)
		for _, dep := range dependencies {
			if dep.storeName == storeName && dep.key == key {
				UpdateConditionBindings(c, conditionID, content.conditionStr)
				break
			}
		}
	}
}

func UpdateStoreBindings(c *HTMLComponent, storeName, key string, newValue interface{}) {
	document := js.Global().Get("document")
	var element js.Value
	if c.ID == "" {
		element = document.Call("getElementById", "app")
	} else {
		element = document.Call("querySelector", fmt.Sprintf("[data-component-id='%s']", c.ID))
	}
	if element.IsNull() || element.IsUndefined() {
		return
	}

	selector := fmt.Sprintf(`[data-store="%s.%s"]`, storeName, key)
	nodes := element.Call("querySelectorAll", selector)
	for i := 0; i < nodes.Length(); i++ {
		node := nodes.Index(i)
		node.Set("innerHTML", fmt.Sprintf("%v", newValue))
	}

	inputSelector := fmt.Sprintf(`input[value="@store:%s.%s:w"], select[value="@store:%s.%s:w"], textarea[value="@store:%s.%s:w"]`, storeName, key, storeName, key, storeName, key)
	inputs := element.Call("querySelectorAll", inputSelector)
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		input.Set("value", fmt.Sprintf("%v", newValue))
	}

	UpdateConditionsForStoreVariable(c, storeName, key)
}

func replaceForeachPlaceholders(template string, c *HTMLComponent) string {
	foreachRegex := regexp.MustCompile(`@foreach:(\S+)\s+as\s+(\w+)([\s\S]*?)@endforeach`)
	return foreachRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := foreachRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		collectionExpr := parts[1]
		itemAlias := parts[2]
		loopContent := parts[3]

		var collection []interface{}

		if strings.HasPrefix(collectionExpr, "store:") {
			storeParts := strings.Split(strings.TrimPrefix(collectionExpr, "store:"), ".")
			if len(storeParts) == 2 {
				storeName, key := storeParts[0], storeParts[1]
				store := GlobalStoreManager.GetStore(storeName)
				if store != nil {
					if col, ok := store.Get(key).([]interface{}); ok {
						collection = col

						unsubscribe := store.OnChange(key, func(newValue interface{}) {
							UpdateDOM(c.ID, c.Render())
						})
						c.unsubscribes = append(c.unsubscribes, unsubscribe)
					} else {
						return match
					}
				} else {
					return match
				}
			} else {
				return match
			}
		} else if value, exists := c.Props[collectionExpr]; exists {
			if col, ok := value.([]interface{}); ok {
				collection = col
			} else {
				return match
			}
		} else {
			return match
		}

		var result strings.Builder

		for _, item := range collection {
			iterContent := loopContent

			if itemMap, ok := item.(map[string]interface{}); ok {
				fieldRegex := regexp.MustCompile(fmt.Sprintf(`@prop:%s\.(\w+)`, itemAlias))
				iterContent = fieldRegex.ReplaceAllStringFunc(iterContent, func(fieldMatch string) string {
					fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
					if len(fieldParts) == 2 {
						fieldName := fieldParts[1]
						if fieldValue, exists := itemMap[fieldName]; exists {
							return fmt.Sprintf("%v", fieldValue)
						}
					}
					return fieldMatch
				})
			} else {
				iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", itemAlias), fmt.Sprintf("%v", item))
			}

			result.WriteString(iterContent)
		}

		return result.String()
	})
}
