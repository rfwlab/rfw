//go:build js && wasm

package core

import (
	"crypto/sha1"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

// AST structures for template parsing
type Node interface {
	Render(c *HTMLComponent) string
}

type TextNode struct {
	Text string
}

func (t *TextNode) Render(c *HTMLComponent) string { return t.Text }

type ConditionalBranch struct {
	Condition string // empty for @else
	Nodes     []Node
}

type ConditionalNode struct {
	Branches []ConditionalBranch
}

// ConditionContent stores rendered content for each branch of a conditional block
type ConditionalBranchContent struct {
	Condition string
	Content   string
}

type ConditionContent struct {
	Branches []ConditionalBranchContent
}

type ConditionDependency struct {
	module    string
	storeName string
	key       string
}

// Render evaluates the conditional branches and renders the appropriate content.
func (cn *ConditionalNode) Render(c *HTMLComponent) string {
	var conditions []string
	for _, br := range cn.Branches {
		conditions = append(conditions, br.Condition)
	}
	conditionID := fmt.Sprintf("cond-%x", sha1.Sum([]byte(strings.Join(conditions, "|"))))

	var content ConditionContent
	var chosen string
	for _, br := range cn.Branches {
		var sb strings.Builder
		for _, n := range br.Nodes {
			sb.WriteString(n.Render(c))
		}
		branchContent := sb.String()
		content.Branches = append(content.Branches, ConditionalBranchContent{Condition: br.Condition, Content: branchContent})

		if br.Condition != "" {
			result, dependencies := evaluateCondition(br.Condition, c)
			for _, dep := range dependencies {
				module, storeName, key := dep.module, dep.storeName, dep.key
				store := state.GlobalStoreManager.GetStore(module, storeName)
				if store != nil {
					unsubscribe := store.OnChange(key, func(newValue interface{}) {
						updateConditionBindings(c, conditionID)
					})
					c.unsubscribes.Add(unsubscribe)
				}
			}
			if chosen == "" && result {
				chosen = branchContent
			}
		} else if chosen == "" {
			chosen = branchContent
		}
	}

	c.conditionContents[conditionID] = content
	return fmt.Sprintf(`<div data-condition="%s">%s</div>`, conditionID, chosen)
}

func replaceIncludePlaceholders(c *HTMLComponent, renderedTemplate string) string {
	includeRegex := regexp.MustCompile(`@include:(\w+)`)
	return includeRegex.ReplaceAllStringFunc(renderedTemplate, func(match string) string {
		name := includeRegex.FindStringSubmatch(match)[1]
		if dep, ok := c.Dependencies[name]; ok {
			return dep.Render()
		}
		return match
	})
}

func extractSlotContents(template string, c *HTMLComponent) string {
	slotRegex := regexp.MustCompile(`@slot:(\w+)(?:\.(\w+))?([\s\S]*?)@endslot`)
	return slotRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := slotRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		depName := parts[1]
		slotName := parts[2]
		if slotName == "" {
			slotName = "default"
		}
		content := parts[3]
		if dep, ok := c.Dependencies[depName]; ok {
			dep.SetSlots(map[string]string{slotName: content})
			return ""
		}
		return match
	})
}

func replaceSlotPlaceholders(template string, c *HTMLComponent) string {
	slotRegex := regexp.MustCompile(`@slot(?::(\w+))?([\s\S]*?)@endslot`)
	return slotRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := slotRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		slotName := parts[1]
		if slotName == "" {
			slotName = "default"
		}
		fallback := parts[2]
		if content, ok := c.Slots[slotName]; ok {
			return content
		}
		return fallback
	})
}

func replaceStorePlaceholders(template string, c *HTMLComponent) string {
	storeRegex := regexp.MustCompile(`@store:(\w+)\.(\w+)\.(\w+)(:w)?`)
	return storeRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := storeRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		module := parts[1]
		storeName := parts[2]
		key := parts[3]
		isWriteable := len(parts) == 5 && parts[4] == ":w"

		store := state.GlobalStoreManager.GetStore(module, storeName)
		if store != nil {
			value := store.Get(key)
			if value == nil {
				value = ""
			}

			unsubscribe := store.OnChange(key, func(newValue interface{}) {
				updateStoreBindings(c, module, storeName, key, newValue)
			})
			c.unsubscribes.Add(unsubscribe)

			if isWriteable {
				return match
			} else {
				return fmt.Sprintf(`<span data-store="%s.%s.%s">%v</span>`, module, storeName, key, value)
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

func replaceEventHandlers(template string) string {
	eventRegex := regexp.MustCompile(`@on:(\w+(?:\.\w+)*)=(?:"([^"]+)"|'([^']+)')`)
	return eventRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := eventRegex.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		handler := parts[2]
		if handler == "" {
			handler = parts[3]
		}
		fullEvent := parts[1]
		eventParts := strings.Split(fullEvent, ".")
		event := eventParts[0]
		modifiers := []string{}
		if len(eventParts) > 1 {
			modifiers = eventParts[1:]
		}
		attr := fmt.Sprintf("data-on-%s=\"%s\"", event, handler)
		if len(modifiers) > 0 {
			attr += fmt.Sprintf(" data-on-%s-modifiers=\"%s\"", event, strings.Join(modifiers, ","))
		}
		return attr
	})
}

// replaceRtIsAttributes scans the template for elements decorated with the
// `rt-is` attribute. The attribute's value identifies a component registered in
// the ComponentRegistry. Matching elements are replaced with an @include
// placeholder so standard include processing can render the referenced
// component and manage its lifecycle.
func replaceRtIsAttributes(template string, c *HTMLComponent) string {
	re := regexp.MustCompile(`<([a-zA-Z0-9]+)([^>]*)rt-is="([^"]+)"[^>]*/?>`)
	idx := 0
	return re.ReplaceAllStringFunc(template, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		name := parts[3]
		comp := LoadComponent(name)
		if comp == nil {
			return match
		}
		placeholder := fmt.Sprintf("rtis-%s-%d", name, idx)
		idx++
		c.AddDependency(placeholder, comp)
		return fmt.Sprintf("@include:%s", placeholder)
	})
}

// parseTemplate parses the template string into an AST of nodes.
func parseTemplate(template string) ([]Node, error) {
	lines := strings.Split(template, "\n")
	idx := 0
	return parseBlock(lines, &idx)
}

func parseBlock(lines []string, idx *int) ([]Node, error) {
	var nodes []Node
	for *idx < len(lines) {
		line := lines[*idx]
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "@if:"):
			cond := trimmed
			*idx++
			n, err := parseConditional(lines, idx, cond)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, n)
		case strings.HasPrefix(trimmed, "@else-if:"), trimmed == "@else", trimmed == "@endif":
			return nodes, nil
		default:
			nodes = append(nodes, &TextNode{Text: line + "\n"})
			*idx++
		}
	}
	return nodes, nil
}

func parseConditional(lines []string, idx *int, firstCond string) (Node, error) {
	node := &ConditionalNode{}
	children, err := parseBlock(lines, idx)
	if err != nil {
		return nil, err
	}
	node.Branches = append(node.Branches, ConditionalBranch{Condition: firstCond, Nodes: children})

	for *idx < len(lines) {
		trimmed := strings.TrimSpace(lines[*idx])
		switch {
		case strings.HasPrefix(trimmed, "@else-if:"):
			cond := trimmed
			*idx++
			children, err := parseBlock(lines, idx)
			if err != nil {
				return nil, err
			}
			node.Branches = append(node.Branches, ConditionalBranch{Condition: cond, Nodes: children})
		case trimmed == "@else":
			*idx++
			children, err := parseBlock(lines, idx)
			if err != nil {
				return nil, err
			}
			node.Branches = append(node.Branches, ConditionalBranch{Condition: "", Nodes: children})
		case trimmed == "@endif":
			*idx++
			return node, nil
		default:
			*idx++
		}
	}
	return node, nil
}

// replaceConditionals parses conditionals using the AST and renders them.
func replaceConditionals(template string, c *HTMLComponent) string {
	nodes, err := parseTemplate(template)
	if err != nil {
		return template
	}
	var sb strings.Builder
	for _, n := range nodes {
		sb.WriteString(n.Render(c))
	}
	return sb.String()
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
	leftSide = strings.Replace(leftSide, "@else-if:", "", 1)
	expectedValue := strings.ReplaceAll(conditionParts[1], `"`, "")
	expectedValue = strings.TrimSpace(expectedValue)

	fmt.Printf("Left side: '%s', Expected value: '%s'\n", leftSide, expectedValue)

	dependencies := []ConditionDependency{}

	if strings.HasPrefix(leftSide, "store:") {
		storeParts := strings.Split(strings.TrimPrefix(leftSide, "store:"), ".")
		if len(storeParts) == 3 {
			module, storeName, key := storeParts[0], storeParts[1], storeParts[2]
			fmt.Printf("Dependency detected: Module '%s', Store '%s', Key '%s'\n", module, storeName, key)
			store := state.GlobalStoreManager.GetStore(module, storeName)
			if store != nil {
				dependencies = append(dependencies, ConditionDependency{module, storeName, key})
				actualValue := fmt.Sprintf("%v", store.Get(key))
				fmt.Printf("Actual value from store: '%s'\n", actualValue)
				return actualValue == expectedValue, dependencies
			} else {
				fmt.Printf("Store '%s' in module '%s' not found.\n", storeName, module)
			}
		} else {
			fmt.Println("Store parts length is not 3.")
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

func updateStoreBindings(c *HTMLComponent, module, storeName, key string, newValue interface{}) {
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

	selector := fmt.Sprintf(`[data-store="%s.%s.%s"]`, module, storeName, key)
	nodes := element.Call("querySelectorAll", selector)
	for i := 0; i < nodes.Length(); i++ {
		node := nodes.Index(i)
		node.Set("innerHTML", fmt.Sprintf("%v", newValue))
	}

	inputSelector := fmt.Sprintf(`input[value="@store:%s.%s.%s:w"], select[value="@store:%s.%s.%s:w"], textarea[value="@store:%s.%s.%s:w"]`, module, storeName, key, module, storeName, key, module, storeName, key)
	inputs := element.Call("querySelectorAll", inputSelector)
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		input.Set("value", fmt.Sprintf("%v", newValue))
	}

	updateConditionsForStoreVariable(c, module, storeName, key)
}

func insertDataKey(content string, key interface{}) string {
	tagRegex := regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9-]*)`)
	loc := tagRegex.FindStringSubmatchIndex(content)
	if loc == nil {
		return content
	}
	return content[:loc[1]] + fmt.Sprintf(` data-key="%v"`, key) + content[loc[1]:]
}

func resolveNumber(expr string, c *HTMLComponent) (int, error) {
	if n, err := strconv.Atoi(expr); err == nil {
		return n, nil
	}
	if strings.HasPrefix(expr, "store:") {
		parts := strings.Split(strings.TrimPrefix(expr, "store:"), ".")
		if len(parts) == 3 {
			module, storeName, key := parts[0], parts[1], parts[2]
			store := state.GlobalStoreManager.GetStore(module, storeName)
			if store != nil {
				if val := store.Get(key); val != nil {
					unsubscribe := store.OnChange(key, func(newValue interface{}) {
						dom.UpdateDOM(c.ID, c.Render())
					})
					c.unsubscribes.Add(unsubscribe)
					switch v := val.(type) {
					case int:
						return v, nil
					case float64:
						return int(v), nil
					case string:
						return strconv.Atoi(v)
					}
				}
			}
		}
	}
	if val, ok := c.Props[expr]; ok {
		switch v := val.(type) {
		case int:
			return v, nil
		case float64:
			return int(v), nil
		case string:
			return strconv.Atoi(v)
		}
	}
	return 0, fmt.Errorf("invalid number")
}

func replaceForPlaceholders(template string, c *HTMLComponent) string {
	forRegex := regexp.MustCompile(`@for:(\w+(?:,\w+)?)\s+in\s+(\S+)([\s\S]*?)@endfor`)
	return forRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := forRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		varsPart := parts[1]
		expr := parts[2]
		loopContent := parts[3]

		aliases := strings.Split(varsPart, ",")
		for i := range aliases {
			aliases[i] = strings.TrimSpace(aliases[i])
		}

		if strings.Contains(expr, "..") {
			rangeParts := strings.Split(expr, "..")
			if len(rangeParts) != 2 {
				return match
			}
			start, err := resolveNumber(rangeParts[0], c)
			if err != nil {
				return match
			}
			end, err := resolveNumber(rangeParts[1], c)
			if err != nil {
				return match
			}
			var result strings.Builder
			for i := start; i <= end; i++ {
				iter := strings.ReplaceAll(loopContent, fmt.Sprintf("@prop:%s", aliases[0]), fmt.Sprintf("%d", i))
				iter = insertDataKey(iter, i)
				result.WriteString(iter)
			}
			return result.String()
		}

		var collection interface{}
		if strings.HasPrefix(expr, "store:") {
			storeParts := strings.Split(strings.TrimPrefix(expr, "store:"), ".")
			if len(storeParts) == 3 {
				module, storeName, key := storeParts[0], storeParts[1], storeParts[2]
				store := state.GlobalStoreManager.GetStore(module, storeName)
				if store != nil {
					collection = store.Get(key)
					unsubscribe := store.OnChange(key, func(newValue interface{}) {
						dom.UpdateDOM(c.ID, c.Render())
					})
					c.unsubscribes.Add(unsubscribe)
				} else {
					return match
				}
			} else {
				return match
			}
		} else if val, ok := c.Props[expr]; ok {
			collection = val
		} else {
			return match
		}

		switch col := collection.(type) {
		case []interface{}:
			var result strings.Builder
			alias := aliases[0]
			for idx, item := range col {
				iterContent := loopContent
				if itemMap, ok := item.(map[string]interface{}); ok {
					fieldRegex := regexp.MustCompile(fmt.Sprintf(`@prop:%s\.(\w+)`, alias))
					iterContent = fieldRegex.ReplaceAllStringFunc(iterContent, func(fieldMatch string) string {
						fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
						if len(fieldParts) == 2 {
							if fieldValue, exists := itemMap[fieldParts[1]]; exists {
								return fmt.Sprintf("%v", fieldValue)
							}
						}
						return fieldMatch
					})
				} else {
					iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", alias), fmt.Sprintf("%v", item))
				}
				iterContent = insertDataKey(iterContent, idx)
				result.WriteString(iterContent)
			}
			return result.String()
		case map[string]interface{}:
			keyAlias := aliases[0]
			valAlias := keyAlias
			if len(aliases) > 1 {
				valAlias = aliases[1]
			}
			keys := make([]string, 0, len(col))
			for k := range col {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var result strings.Builder
			for _, k := range keys {
				v := col[k]
				iterContent := strings.ReplaceAll(loopContent, fmt.Sprintf("@prop:%s", keyAlias), k)
				if len(aliases) > 1 {
					if vMap, ok := v.(map[string]interface{}); ok {
						fieldRegex := regexp.MustCompile(fmt.Sprintf(`@prop:%s\.(\w+)`, valAlias))
						iterContent = fieldRegex.ReplaceAllStringFunc(iterContent, func(fieldMatch string) string {
							fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
							if len(fieldParts) == 2 {
								if fieldValue, exists := vMap[fieldParts[1]]; exists {
									return fmt.Sprintf("%v", fieldValue)
								}
							}
							return fieldMatch
						})
					} else {
						iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", valAlias), fmt.Sprintf("%v", v))
					}
				}
				iterContent = insertDataKey(iterContent, k)
				result.WriteString(iterContent)
			}
			return result.String()
		default:
			return match
		}
	})
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
			if len(storeParts) == 3 {
				module, storeName, key := storeParts[0], storeParts[1], storeParts[2]
				store := state.GlobalStoreManager.GetStore(module, storeName)
				if store != nil {
					if col, ok := store.Get(key).([]interface{}); ok {
						collection = col

						unsubscribe := store.OnChange(key, func(newValue interface{}) {
							dom.UpdateDOM(c.ID, c.Render())
						})
						c.unsubscribes.Add(unsubscribe)
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

func updateConditionBindings(c *HTMLComponent, conditionID string) {
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

	conditionContent := c.conditionContents[conditionID]
	var newContent string
	for _, br := range conditionContent.Branches {
		if br.Condition == "" {
			if newContent == "" {
				newContent = br.Content
			}
			continue
		}
		result, _ := evaluateCondition(br.Condition, c)
		if result {
			newContent = br.Content
			break
		}
	}

	node.Set("innerHTML", newContent)

	dom.BindStoreInputs(node)
}

func updateConditionsForStoreVariable(c *HTMLComponent, module, storeName, key string) {
	for conditionID, content := range c.conditionContents {
		for _, br := range content.Branches {
			if br.Condition == "" {
				continue
			}
			dependencies, _ := getConditionDependencies(br.Condition)
			for _, dep := range dependencies {
				if dep.module == module && dep.storeName == storeName && dep.key == key {
					updateConditionBindings(c, conditionID)
					break
				}
			}
		}
	}
}

func getConditionDependencies(condition string) ([]ConditionDependency, error) {
	conditionParts := strings.Split(condition, "==")
	if len(conditionParts) != 2 {
		return nil, fmt.Errorf("Invalid condition format")
	}

	leftSide := strings.TrimSpace(conditionParts[0])
	leftSide = strings.Replace(leftSide, "@if:", "", 1)
	leftSide = strings.Replace(leftSide, "@else-if:", "", 1)

	dependencies := []ConditionDependency{}

	if strings.HasPrefix(leftSide, "store:") {
		storeParts := strings.Split(strings.TrimPrefix(leftSide, "store:"), ".")
		if len(storeParts) == 3 {
			module, storeName, key := storeParts[0], storeParts[1], storeParts[2]
			dependencies = append(dependencies, ConditionDependency{module, storeName, key})
		}
	}

	return dependencies, nil
}
