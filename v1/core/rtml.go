//go:build js && wasm

package core

import (
	"crypto/sha1"
	"fmt"
	"html"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

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
	signal    string
}

type ForeachConfig struct {
	Expr      string
	ItemAlias string
	Content   string
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
				if dep.signal != "" {
					if prop, ok := c.Props[dep.signal]; ok {
						if sig, ok := prop.(interface{ Read() any }); ok {
							dom.RegisterSignal(c.ID, dep.signal, sig)
							unsub := state.Effect(func() func() {
								sig.Read()
								updateConditionBindings(c, conditionID)
								return nil
							})
							c.unsubscribes.Add(unsub)
						}
					}
					continue
				}
				module, storeName, key := dep.module, dep.storeName, dep.key
				store := state.GlobalStoreManager.GetStore(module, storeName)
				if store != nil {
					unsubscribe := store.OnChange(key, func(newValue any) {
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
	includeRegex := regexp.MustCompile(`@include:([\w-]+)`)
	return includeRegex.ReplaceAllStringFunc(renderedTemplate, func(match string) string {
		name := includeRegex.FindStringSubmatch(match)[1]
		if dep, ok := c.Dependencies[name]; ok {
			return dep.Render()
		}
		if DevMode {
			Log().Warn("component %s missing dependency '%s'", c.Name, name)
		}
		return match
	})
}

// replaceComponentIncludes scans for @include directives that supply inline
// props using the syntax @include:Component:{key:"value"}. Matching includes
// are replaced with standard @include placeholders after instantiating the
// component and registering it as a dependency.
func replaceComponentIncludes(template string, c *HTMLComponent) string {
	idx := 0

	// Handle includes that may be wrapped in <p> tags produced by Markdown
	// renderers as well as bare @include directives.
	patterns := []string{
		`<p>@include:([\w-]+):\{([^}]*)\}</p>`,
		`@include:([\w-]+):\{([^}]*)\}`,
	}

	for _, pat := range patterns {
		re := regexp.MustCompile(pat)
		template = re.ReplaceAllStringFunc(template, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) < 3 {
				return match
			}
			name := parts[1]
			propStr := html.UnescapeString(parts[2])
			comp := LoadComponent(name)
			if comp == nil {
				if DevMode {
					Log().Warn("include referenced unknown component '%s'", name)
				}
				return match
			}
			props := map[string]any{}
			propRe := regexp.MustCompile(`(\w+):"([^"]*)"`)
			for _, m := range propRe.FindAllStringSubmatch(propStr, -1) {
				props[m[1]] = m[2]
			}
			if hc, ok := comp.(*HTMLComponent); ok {
				hc.Props = props
				hc.ID = generateComponentID(hc.Name, hc.Props)
			}
			placeholder := fmt.Sprintf("inc-%s-%d", name, idx)
			idx++
			c.AddDependency(placeholder, comp)
			return "@include:" + placeholder
		})
	}

	return template
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
			dep.SetSlots(map[string]any{slotName: content})
			return ""
		}
		if DevMode {
			Log().Warn("component %s missing dependency '%s' for slot '%s'", c.Name, depName, slotName)
		}
		return match
	})
}

func replaceSlotPlaceholders(template string, c *HTMLComponent) string {
	slotRegex := regexp.MustCompile(`@slot(?::(\w+))?([\s\S]*?)@endslot`)
	idx := 0
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
			switch v := content.(type) {
			case string:
				return v
			case Component:
				placeholder := fmt.Sprintf("slot-%s-%d", slotName, idx)
				idx++
				c.AddDependency(placeholder, v)
				return fmt.Sprintf("@include:%s", placeholder)
			default:
				return fallback
			}
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

			unsubscribe := store.OnChange(key, func(newValue any) {
				updateStoreBindings(c, module, storeName, key, newValue)
			})
			c.unsubscribes.Add(unsubscribe)

			if isWriteable {
				return match
			} else {
				return fmt.Sprintf(`<span data-store="%s.%s.%s">%v</span>`, module, storeName, key, value)
			}
		}
		if DevMode {
			Log().Warn("store %s.%s not found for key '%s' in component %s", module, storeName, key, c.Name)
		}
		return match
	})
}

func replaceSignalPlaceholders(template string, c *HTMLComponent) string {
	sigRegex := regexp.MustCompile(`@signal:(\w+)(:w)?`)
	return sigRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := sigRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		name := parts[1]
		isWriteable := len(parts) == 3 && parts[2] == ":w"
		if prop, ok := c.Props[name]; ok {
			if sig, ok := prop.(interface{ Read() any }); ok {
				dom.RegisterSignal(c.ID, name, sig)
				val := sig.Read()
				unsub := state.Effect(func() func() {
					v := sig.Read()
					updateSignalBindings(c, name, v)
					return nil
				})
				c.unsubscribes.Add(unsub)
				if isWriteable {
					return match
				}
				return fmt.Sprintf(`<span data-signal="%s">%v</span>`, name, val)
			}
		}
		if DevMode {
			Log().Warn("signal '%s' not found in component %s", name, c.Name)
		}
		return match
	})
}

func replacePropPlaceholders(template string, c *HTMLComponent) string {
	propRegex := regexp.MustCompile(`@prop:(\w+)`)
	idx := 0
	return propRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := propRegex.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		propName := parts[1]
		if value, exists := c.Props[propName]; exists {
			switch v := value.(type) {
			case Component:
				placeholder := fmt.Sprintf("prop-%s-%d", propName, idx)
				idx++
				c.AddDependency(placeholder, v)
				return fmt.Sprintf("@include:%s", placeholder)
			default:
				return fmt.Sprintf("%v", v)
			}
		}
		if DevMode {
			Log().Warn("component %s missing prop '%s'", c.Name, propName)
		}
		return match
	})
}

func replacePluginPlaceholders(template string) string {
	varRegex := regexp.MustCompile(`\{plugin:(\w+)\.(\w+)\}`)
	template = varRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := varRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		plug, name := parts[1], parts[2]
		if v, ok := getRTMLVar(plug, name); ok {
			return fmt.Sprintf("%v", v)
		}
		if DevMode {
			Log().Warn("plugin variable %s.%s not found", plug, name)
		}
		return match
	})
	cmdRegex := regexp.MustCompile(`@plugin:(\w+)\.(\w+)([\s>/])`)
	template = cmdRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := cmdRegex.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		plug, name, suffix := parts[1], parts[2], parts[3]
		return fmt.Sprintf(`data-plugin-cmd="%s.%s"%s`, plug, name, suffix)
	})
	return template
}

func replaceHostPlaceholders(template string, c *HTMLComponent) string {
	varRegex := regexp.MustCompile(`\{h:(\w+)\}`)
	template = varRegex.ReplaceAllStringFunc(template, func(match string) string {
		name := varRegex.FindStringSubmatch(match)[1]
		c.hostVars = append(c.hostVars, name)
		return fmt.Sprintf(`<span data-host-var="%s"></span>`, name)
	})
	cmdRegex := regexp.MustCompile(`@h:(\w+)`)
	template = cmdRegex.ReplaceAllStringFunc(template, func(match string) string {
		name := cmdRegex.FindStringSubmatch(match)[1]
		c.hostCmds = append(c.hostCmds, name)
		return fmt.Sprintf(`data-host-cmd="%s"`, name)
	})
	return template
}

func replaceEventHandlers(template string) string {
	// Match event directives ensuring they are terminated by whitespace,
	// a self-closing slash or the end of the tag. The terminating
	// character is captured so it can be preserved in the replacement.
	eventRegex := regexp.MustCompile(`@(on:)?(\w+(?:\.\w+)*):(\w+)([\s>/])`)
	return eventRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := eventRegex.FindStringSubmatch(match)
		if len(parts) != 5 {
			return match
		}
		fullEvent := parts[2]
		handler := parts[3]
		suffix := parts[4]
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
		return attr + suffix
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
			if DevMode {
				Log().Warn("rt-is referenced unknown component '%s'", name)
			}
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
	Log().Debug("Evaluating condition: '%s'", condition)
	conditionParts := strings.Split(condition, "==")
	if len(conditionParts) != 2 {
		Log().Debug("Condition format is invalid. Expected '=='.")
		return false, nil
	}

	leftSide := strings.TrimSpace(conditionParts[0])
	leftSide = strings.Replace(leftSide, "@if:", "", 1)
	leftSide = strings.Replace(leftSide, "@else-if:", "", 1)
	expectedValue := strings.ReplaceAll(conditionParts[1], `"`, "")
	expectedValue = strings.TrimSpace(expectedValue)

	Log().Debug("Left side: '%s', Expected value: '%s'", leftSide, expectedValue)

	dependencies := []ConditionDependency{}

	if strings.HasPrefix(leftSide, "store:") {
		storeParts := strings.Split(strings.TrimPrefix(leftSide, "store:"), ".")
		if len(storeParts) == 3 {
			module, storeName, key := storeParts[0], storeParts[1], storeParts[2]
			Log().Debug("Dependency detected: Module '%s', Store '%s', Key '%s'", module, storeName, key)
			store := state.GlobalStoreManager.GetStore(module, storeName)
			if store != nil {
				dependencies = append(dependencies, ConditionDependency{module: module, storeName: storeName, key: key})
				actualValue := fmt.Sprintf("%v", store.Get(key))
				Log().Debug("Actual value from store: '%s'", actualValue)
				return actualValue == expectedValue, dependencies
			} else {
				Log().Debug("Store '%s' in module '%s' not found.", storeName, module)
			}
		} else {
			Log().Debug("Store parts length is not 3.")
		}
	}

	if strings.HasPrefix(leftSide, "signal:") {
		sigName := strings.TrimPrefix(leftSide, "signal:")
		if prop, ok := c.Props[sigName]; ok {
			if sig, ok := prop.(interface{ Read() any }); ok {
				dependencies = append(dependencies, ConditionDependency{signal: sigName})
				actualValue := fmt.Sprintf("%v", sig.Read())
				return actualValue == expectedValue, dependencies
			}
		}
	}

	if strings.HasPrefix(leftSide, "prop:") {
		propName := strings.TrimPrefix(leftSide, "prop:")
		if value, exists := c.Props[propName]; exists {
			actualValue := fmt.Sprintf("%v", value)
			Log().Debug("Actual value from props: '%s'", actualValue)
			return actualValue == expectedValue, dependencies
		}
	}

	Log().Debug("No dependencies detected.")
	return false, dependencies
}

func updateStoreBindings(c *HTMLComponent, module, storeName, key string, newValue any) {
	doc := dom.Doc()
	var element dom.Element
	if c.ID == "" {
		element = doc.ByID("app")
	} else {
		element = doc.Query(fmt.Sprintf("[data-component-id='%s']", c.ID))
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

	placeholder := fmt.Sprintf("@store:%s.%s.%s:w", module, storeName, key)

	// Update value-based inputs and selects
	inputSelector := fmt.Sprintf(`input[value="%s"], select[value="%s"]`, placeholder, placeholder)
	inputs := element.Call("querySelectorAll", inputSelector)
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		input.Set("value", fmt.Sprintf("%v", newValue))
	}

	// Update checkboxes bound via checked attribute
	checkedSelector := fmt.Sprintf(`input[checked="%s"]`, placeholder)
	checks := element.Call("querySelectorAll", checkedSelector)
	for i := 0; i < checks.Length(); i++ {
		chk := checks.Index(i)
		switch v := newValue.(type) {
		case bool:
			chk.Set("checked", v)
		case string:
			chk.Set("checked", strings.ToLower(v) == "true")
		default:
			chk.Set("checked", newValue != nil)
		}
	}

	// Update textareas where placeholder is in content
	textareas := element.Call("querySelectorAll", "textarea")
	for i := 0; i < textareas.Length(); i++ {
		ta := textareas.Index(i)
		if ta.Get("value").String() == placeholder {
			ta.Set("value", fmt.Sprintf("%v", newValue))
		}
	}

	updateConditionsForStoreVariable(c, module, storeName, key)
}

func updateSignalBindings(c *HTMLComponent, name string, newValue any) {
	doc := dom.Doc()
	var element dom.Element
	if c.ID == "" {
		element = doc.ByID("app")
	} else {
		element = doc.Query(fmt.Sprintf("[data-component-id='%s']", c.ID))
	}
	if element.IsNull() || element.IsUndefined() {
		return
	}

	selector := fmt.Sprintf(`[data-signal="%s"]`, name)
	nodes := element.Call("querySelectorAll", selector)
	for i := 0; i < nodes.Length(); i++ {
		node := nodes.Index(i)
		node.Set("innerHTML", fmt.Sprintf("%v", newValue))
	}

	placeholder := fmt.Sprintf("@signal:%s:w", name)

	// Update value-based inputs and selects
	inputSelector := fmt.Sprintf(`input[value="%s"], select[value="%s"]`, placeholder, placeholder)
	inputs := element.Call("querySelectorAll", inputSelector)
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		input.Set("value", fmt.Sprintf("%v", newValue))
	}

	// Update checkboxes
	checkedSelector := fmt.Sprintf(`input[checked="%s"]`, placeholder)
	checks := element.Call("querySelectorAll", checkedSelector)
	for i := 0; i < checks.Length(); i++ {
		chk := checks.Index(i)
		switch v := newValue.(type) {
		case bool:
			chk.Set("checked", v)
		case string:
			chk.Set("checked", strings.ToLower(v) == "true")
		default:
			chk.Set("checked", newValue != nil)
		}
	}

	// Update textareas with placeholder in content
	textareas := element.Call("querySelectorAll", "textarea")
	for i := 0; i < textareas.Length(); i++ {
		ta := textareas.Index(i)
		if ta.Get("value").String() == placeholder {
			ta.Set("value", fmt.Sprintf("%v", newValue))
		}
	}
}

func insertDataKey(content string, key any) string {
	tagRegex := regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9-]*)`)
	loc := tagRegex.FindStringSubmatchIndex(content)
	if loc == nil {
		return content
	}
	return content[:loc[1]] + fmt.Sprintf(` data-key="%v"`, key) + content[loc[1]:]
}

// replaceConstructors scans for inline constructor tokens inside an element's
// start tag and injects the corresponding data attribute. Supported
// constructors:
//
//	[name]       -> data-ref="name"
//	[key expr]   -> data-key="expr"
//
// Only a single constructor per element is handled.
func replaceConstructors(template string) string {
	re := regexp.MustCompile(`<([a-zA-Z][\w-]*)([^>]*?)\s\[([^\] ]+)(?:\s+([^\]]+))?\]([^>]*)>`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 6 {
			return match
		}
		tag := parts[1]
		before := parts[2]
		name := parts[3]
		param := parts[4]
		after := parts[5]
		attr := ""
		if name == "key" && param != "" {
			attr = fmt.Sprintf(` data-key="%s"`, param)
		} else if strings.HasPrefix(name, "plugin:") {
			attr = fmt.Sprintf(` data-plugin="%s"`, strings.TrimPrefix(name, "plugin:"))
		} else {
			attr = fmt.Sprintf(` data-ref="%s"`, name)
		}
		return fmt.Sprintf("<%s%s%s%s>", tag, before, attr, after)
	})
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
					unsubscribe := store.OnChange(key, func(newValue any) {
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

func legacyReplaceForPlaceholders(template string, c *HTMLComponent) string {
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

		var collection any
		if strings.HasPrefix(expr, "store:") {
			storeParts := strings.Split(strings.TrimPrefix(expr, "store:"), ".")
			if len(storeParts) == 3 {
				module, storeName, key := storeParts[0], storeParts[1], storeParts[2]
				store := state.GlobalStoreManager.GetStore(module, storeName)
				if store != nil {
					collection = store.Get(key)
					unsubscribe := store.OnChange(key, func(newValue any) {
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
		case []any:
			var result strings.Builder
			alias := aliases[0]
			for idx, item := range col {
				iterContent := loopContent
				if itemMap, ok := item.(map[string]any); ok {
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
		case map[string]any:
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
					if vMap, ok := v.(map[string]any); ok {
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

		expr := parts[1]
		alias := parts[2]
		content := parts[3]
		foreachID := fmt.Sprintf("foreach-%x", sha1.Sum([]byte(match)))
		c.foreachContents[foreachID] = ForeachConfig{Expr: expr, ItemAlias: alias, Content: content}

		if strings.HasPrefix(expr, "store:") {
			storeParts := strings.Split(strings.TrimPrefix(expr, "store:"), ".")
			if len(storeParts) == 3 {
				module, storeName, key := storeParts[0], storeParts[1], storeParts[2]
				store := state.GlobalStoreManager.GetStore(module, storeName)
				if store != nil {
					unsubscribe := store.OnChange(key, func(newValue any) {
						updateForeachBindings(c, foreachID)
					})
					c.unsubscribes.Add(unsubscribe)
				}
			}
		} else {
			name := strings.TrimPrefix(expr, "signal:")
			if prop, ok := c.Props[name]; ok {
				if sig, ok := prop.(interface{ Read() any }); ok {
					dom.RegisterSignal(c.ID, name, sig)
					unsub := state.Effect(func() func() {
						sig.Read()
						updateForeachBindings(c, foreachID)
						return nil
					})
					c.unsubscribes.Add(unsub)
				}
			}
		}

		rendered := renderForeachLoop(c, expr, alias, content)
		return fmt.Sprintf(`<div data-foreach="%s">%s</div>`, foreachID, rendered)
	})
}

func renderForeachLoop(c *HTMLComponent, expr, alias, content string) string {
	var collection any
	if strings.HasPrefix(expr, "store:") {
		parts := strings.Split(strings.TrimPrefix(expr, "store:"), ".")
		if len(parts) == 3 {
			module, storeName, key := parts[0], parts[1], parts[2]
			store := state.GlobalStoreManager.GetStore(module, storeName)
			if store != nil {
				collection = store.Get(key)
			}
		}
	} else {
		name := strings.TrimPrefix(expr, "signal:")
		if val, ok := c.Props[name]; ok {
			if sig, ok := val.(interface{ Read() any }); ok {
				collection = sig.Read()
			} else {
				collection = val
			}
		}
	}

	val := reflect.ValueOf(collection)
	if !val.IsValid() {
		return ""
	}
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		var result strings.Builder
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			iter := content
			if itemMap, ok := item.(map[string]any); ok {
				fieldRegex := regexp.MustCompile(fmt.Sprintf(`@%s\\.(\\w+)`, alias))
				iter = fieldRegex.ReplaceAllStringFunc(iter, func(fieldMatch string) string {
					fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
					if len(fieldParts) == 2 {
						if fieldValue, exists := itemMap[fieldParts[1]]; exists {
							return fmt.Sprintf("%v", fieldValue)
						}
					}
					return fieldMatch
				})
				propFieldRegex := regexp.MustCompile(fmt.Sprintf(`@prop:%s\\.(\\w+)`, alias))
				iter = propFieldRegex.ReplaceAllStringFunc(iter, func(fieldMatch string) string {
					fieldParts := propFieldRegex.FindStringSubmatch(fieldMatch)
					if len(fieldParts) == 2 {
						if fieldValue, exists := itemMap[fieldParts[1]]; exists {
							return fmt.Sprintf("%v", fieldValue)
						}
					}
					return fieldMatch
				})
			}
			iter = strings.ReplaceAll(iter, fmt.Sprintf("@%s", alias), fmt.Sprintf("%v", item))
			iter = strings.ReplaceAll(iter, fmt.Sprintf("@prop:%s", alias), fmt.Sprintf("%v", item))
			result.WriteString(iter)
		}
		return result.String()
	case reflect.Map:
		if val.Type().Key().Kind() != reflect.String {
			return ""
		}
		var result strings.Builder
		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		for _, k := range keys {
			v := val.MapIndex(k).Interface()
			iter := content
			if vMap, ok := v.(map[string]any); ok {
				fieldRegex := regexp.MustCompile(fmt.Sprintf(`@%s\\.(\\w+)`, alias))
				iter = fieldRegex.ReplaceAllStringFunc(iter, func(fieldMatch string) string {
					fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
					if len(fieldParts) == 2 {
						if fieldValue, exists := vMap[fieldParts[1]]; exists {
							return fmt.Sprintf("%v", fieldValue)
						}
					}
					return fieldMatch
				})
				propFieldRegex := regexp.MustCompile(fmt.Sprintf(`@prop:%s\\.(\\w+)`, alias))
				iter = propFieldRegex.ReplaceAllStringFunc(iter, func(fieldMatch string) string {
					fieldParts := propFieldRegex.FindStringSubmatch(fieldMatch)
					if len(fieldParts) == 2 {
						if fieldValue, exists := vMap[fieldParts[1]]; exists {
							return fmt.Sprintf("%v", fieldValue)
						}
					}
					return fieldMatch
				})
			}
			iter = strings.ReplaceAll(iter, fmt.Sprintf("@%s", alias), fmt.Sprintf("%v", v))
			iter = strings.ReplaceAll(iter, fmt.Sprintf("@prop:%s", alias), fmt.Sprintf("%v", v))
			result.WriteString(iter)
		}
		return result.String()
	default:
		return ""
	}
}

func updateForeachBindings(c *HTMLComponent, foreachID string) {
	doc := dom.Doc()
	var element dom.Element
	if c.ID == "" {
		element = doc.ByID("app")
	} else {
		element = doc.Query(fmt.Sprintf("[data-component-id='%s']", c.ID))
	}
	if element.IsNull() || element.IsUndefined() {
		return
	}

	selector := fmt.Sprintf(`[data-foreach="%s"]`, foreachID)
	node := element.Call("querySelector", selector)
	if node.IsNull() || node.IsUndefined() {
		return
	}

	cfg := c.foreachContents[foreachID]
	newContent := renderForeachLoop(c, cfg.Expr, cfg.ItemAlias, cfg.Content)
	node.Set("innerHTML", newContent)
	dom.BindStoreInputs(node)
	dom.BindSignalInputs(c.ID, node)
}

func updateConditionBindings(c *HTMLComponent, conditionID string) {
	doc := dom.Doc()
	var element dom.Element
	if c.ID == "" {
		element = doc.ByID("app")
	} else {
		element = doc.Query(fmt.Sprintf("[data-component-id='%s']", c.ID))
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
	dom.BindSignalInputs(c.ID, node)
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
			dependencies = append(dependencies, ConditionDependency{module: module, storeName: storeName, key: key})
		}
	}

	return dependencies, nil
}
