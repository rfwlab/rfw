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

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/rtmlast"
	"github.com/rfwlab/rfw/v2/rtmleval"
	"github.com/rfwlab/rfw/v2/state"
)

var (
	reInclude         = regexp.MustCompile(`@include:([\w-]+)`)
	rePropKV          = regexp.MustCompile(`(\w+):"([^"]*)"`)
	reSlotNamed       = regexp.MustCompile(`@slot:(\w+)(?:\.(\w+))?([\s\S]*?)@endslot`)
	reSlotDefault     = regexp.MustCompile(`@slot(?::(\w+))?([\s\S]*?)@endslot`)
	reStore           = regexp.MustCompile(`@store:(\w+)\.(\w+)\.(\w+)(:w)?`)
	reSignal          = regexp.MustCompile(`@signal:(\w+)(:w)?`)
	reExpr            = regexp.MustCompile(`@expr:([^<@]+)`)
	reProp            = regexp.MustCompile(`@prop:(\w+)`)
	rePluginVar       = regexp.MustCompile(`\{plugin:(\w+)\.(\w+)\}`)
	rePluginCmd       = regexp.MustCompile(`@plugin:(\w+)\.(\w+)([\s>/])`)
	reHelperVar       = regexp.MustCompile(`\{h:(\w+)\}`)
	reHelperCmd       = regexp.MustCompile(`@h:(\w+)`)
	reEvent           = regexp.MustCompile(`@(on:)?(\w+(?:\.\w+)*):(\w+)([\s>/])`)
	reRtIs            = regexp.MustCompile(`<([a-zA-Z0-9]+)([^>]*)rt-is="([^"]+)"[^>]*/?>`)
	reTagName         = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9-]*)`)
	reConditionalAttr = regexp.MustCompile(`<([a-zA-Z][\w-]*)([^>]*?)\s\[([^\] ]+)(?:\s+([^\]]+))?\]([^>]*)>`)
	reFor             = regexp.MustCompile(`@for:(\w+(?:,\w+)?)\s+in\s+(\S+)([\s\S]*?)@endfor`)
	reForeach         = regexp.MustCompile(`@foreach:(\S+)\s+as\s+(\w+)([\s\S]*?)@endforeach`)
	depRegex          = regexp.MustCompile(`(?:store:\w+\.\w+\.\w+|signal:\w+|prop:\w+|\w+(?:\.\w+)*)`)
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
			result, _ := evaluateCondition(br.Condition, c)
			if chosen == "" && result {
				chosen = branchContent
			}
		} else if chosen == "" {
			chosen = branchContent
		}
	}

	c.conditionContents[conditionID] = content

	unsub := state.Effect(func() func() {
		for _, br := range cn.Branches {
			if br.Condition != "" {
				evaluateCondition(br.Condition, c)
			}
		}
		updateConditionBindings(c, conditionID)
		return nil
	})
	c.unsubscribes.Add(unsub)

	return fmt.Sprintf(`<div data-condition="%s">%s</div>`, conditionID, chosen)
}

func replaceIncludePlaceholders(c *HTMLComponent, renderedTemplate string) string {
	includeRegex := reInclude
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
			propRe := rePropKV
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
	slotRegex := reSlotNamed
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
	slotRegex := reSlotNamed
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
	storeRegex := reStore
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
	sigRegex := reSignal
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

func replaceExprPlaceholders(template string, c *HTMLComponent) string {
	exprRegex := reExpr
	idx := 0
	return exprRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := exprRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		exprStr := strings.TrimSpace(parts[1])
		exprID := fmt.Sprintf("expr-%d", idx)
		idx++

		astExpr := rtmlast.ParseExpr(exprStr)
		initialVal := evalASTExpr(astExpr, c)
		c.exprContents[exprID] = exprToString(astExpr)

		unsub := state.Effect(func() func() {
			newVal := evalASTExpr(astExpr, c)
			updateExprBindings(c, exprID, newVal)
			return nil
		})
		c.unsubscribes.Add(unsub)

		return fmt.Sprintf(`<span data-expr="%s">%v</span>`, exprID, initialVal)
	})
}

func replaceExprInClassAttr(template string, c *HTMLComponent) string {
	classRe := regexp.MustCompile(`class="([^"]*@expr:[^"]*)"`)
	idx := 0
	result := classRe.ReplaceAllStringFunc(template, func(match string) string {
		parts := classRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		classVal := parts[1]

		exprInAttrRe := regexp.MustCompile(`@expr:((?:[^"<@]|'[^']*')+)`)
		var exprIDs []string
		newClassVal := exprInAttrRe.ReplaceAllStringFunc(classVal, func(exprMatch string) string {
			eparts := exprInAttrRe.FindStringSubmatch(exprMatch)
			if len(eparts) < 2 {
				return exprMatch
			}
			exprStr := strings.TrimSpace(eparts[1])
			if exprStr == "" || len(exprStr) == 1 && (exprStr[0] == '\'' || exprStr[0] == '"') {
				return exprMatch
			}
			exprID := fmt.Sprintf("class-expr-%d", idx)
			idx++
			exprIDs = append(exprIDs, exprID)

			astExpr := rtmlast.ParseExpr(exprStr)
			initialVal := evalASTExpr(astExpr, c)
			dynamicVal := strings.TrimSpace(fmt.Sprintf("%v", initialVal))
			c.classExprContents[exprID] = dynamicVal
			c.exprContents[exprID] = exprToString(astExpr)

			unsub := state.Effect(func() func() {
				newVal := evalASTExpr(astExpr, c)
				updateClassExprBindings(c, exprID, newVal)
				return nil
			})
			c.unsubscribes.Add(unsub)

			return dynamicVal
		})

		idsStr := strings.Join(exprIDs, " ")
		return fmt.Sprintf(`class="%s" data-expr-class="%s"`, newClassVal, idsStr)
	})
	return result
}

func evalASTExpr(expr rtmlast.Expr, c *HTMLComponent) any {
	switch e := expr.(type) {
	case rtmlast.IdentExpr:
		name := e.Name
		if strings.HasPrefix(name, "store:") {
			parts := strings.Split(strings.TrimPrefix(name, "store:"), ".")
			if len(parts) == 3 {
				store := state.GlobalStoreManager.GetStore(parts[0], parts[1])
				if store != nil {
					return store.Get(parts[2])
				}
			}
			return nil
		}
		if strings.HasPrefix(name, "signal:") {
			sigName := strings.TrimPrefix(name, "signal:")
			if prop, ok := c.Props[sigName]; ok {
				if sig, ok := prop.(interface{ Read() any }); ok {
					return sig.Read()
				}
				return prop
			}
			return nil
		}
		if prop, ok := c.Props[name]; ok {
			if sig, ok := prop.(interface{ Read() any }); ok {
				return sig.Read()
			}
			return prop
		}
		return nil
	case rtmlast.LiteralExpr:
		return e.Value
	case rtmlast.BinaryExpr:
		switch e.Op {
		case rtmlast.OpEq:
			return cmpASTEqual(evalASTExpr(e.LHS, c), evalASTExpr(e.RHS, c))
		case rtmlast.OpNeq:
			return !cmpASTEqual(evalASTExpr(e.LHS, c), evalASTExpr(e.RHS, c))
		case rtmlast.OpAnd:
			return toASTBool(evalASTExpr(e.LHS, c)) && toASTBool(evalASTExpr(e.RHS, c))
		case rtmlast.OpOr:
			return toASTBool(evalASTExpr(e.LHS, c)) || toASTBool(evalASTExpr(e.RHS, c))
		case rtmlast.OpLt, rtmlast.OpGt, rtmlast.OpLte, rtmlast.OpGte:
			return cmpASTValues(evalASTExpr(e.LHS, c), evalASTExpr(e.RHS, c), e.Op)
		default:
			lhs := toASTFloat(evalASTExpr(e.LHS, c))
			rhs := toASTFloat(evalASTExpr(e.RHS, c))
			switch e.Op {
			case rtmlast.OpAdd:
				return lhs + rhs
			case rtmlast.OpSub:
				return lhs - rhs
			case rtmlast.OpMul:
				return lhs * rhs
			case rtmlast.OpDiv:
				if rhs == 0 {
					return 0.0
				}
				return lhs / rhs
			}
		}
	case rtmlast.UnaryExpr:
		val := evalASTExpr(e.Expr, c)
		switch e.Op {
		case rtmlast.UnaryNot:
			return !toASTBool(val)
		case rtmlast.UnaryNeg:
			return -toASTFloat(val)
		}
	case rtmlast.FieldExpr:
		obj := evalASTExpr(e.Obj, c)
		if m, ok := obj.(map[string]any); ok {
			return m[e.Field]
		}
		return nil
	case rtmlast.TernaryExpr:
		if toASTBool(evalASTExpr(e.Cond, c)) {
			return evalASTExpr(e.Then, c)
		}
		return evalASTExpr(e.Else, c)
	}
	return nil
}

func cmpASTEqual(a, b any) bool {
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	}
	af, aok := toASTFloatOk(a)
	bf, bok := toASTFloatOk(b)
	if aok && bok {
		return af == bf
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func cmpASTValues(a, b any, op rtmlast.BinOp) bool {
	l, r := toASTFloat(a), toASTFloat(b)
	switch op {
	case rtmlast.OpLt:
		return l < r
	case rtmlast.OpGt:
		return l > r
	case rtmlast.OpLte:
		return l <= r
	case rtmlast.OpGte:
		return l >= r
	default:
		return false
	}
}

func toASTFloat(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case float32:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func toASTFloatOk(v any) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}

func toASTBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != ""
	case int:
		return val != 0
	case float64:
		return val != 0
	default:
		return v != nil
	}
}

func exprToString(expr rtmlast.Expr) string {
	switch e := expr.(type) {
	case rtmlast.IdentExpr:
		return e.Name
	case rtmlast.LiteralExpr:
		return fmt.Sprintf("%v", e.Value)
	case rtmlast.BinaryExpr:
		return fmt.Sprintf("(%s %s %s)", exprToString(e.LHS), binOpString(e.Op), exprToString(e.RHS))
	case rtmlast.UnaryExpr:
		switch e.Op {
		case rtmlast.UnaryNot:
			return fmt.Sprintf("!%s", exprToString(e.Expr))
		case rtmlast.UnaryNeg:
			return fmt.Sprintf("-%s", exprToString(e.Expr))
		}
	case rtmlast.FieldExpr:
		return fmt.Sprintf("%s.%s", exprToString(e.Obj), e.Field)
	case rtmlast.CallExpr:
		return fmt.Sprintf("%s(%v)", e.Fn, e.Args)
	case rtmlast.TernaryExpr:
		return fmt.Sprintf("%s ? %s : %s", exprToString(e.Cond), exprToString(e.Then), exprToString(e.Else))
	}
	return ""
}

func binOpString(op rtmlast.BinOp) string {
	switch op {
	case rtmlast.OpEq:
		return "=="
	case rtmlast.OpNeq:
		return "!="
	case rtmlast.OpLt:
		return "<"
	case rtmlast.OpGt:
		return ">"
	case rtmlast.OpLte:
		return "<="
	case rtmlast.OpGte:
		return ">="
	case rtmlast.OpAnd:
		return "&&"
	case rtmlast.OpOr:
		return "||"
	case rtmlast.OpAdd:
		return "+"
	case rtmlast.OpSub:
		return "-"
	case rtmlast.OpMul:
		return "*"
	case rtmlast.OpDiv:
		return "/"
	default:
		return "?"
	}
}

func updateExprBindings(c *HTMLComponent, exprID string, newValue any) {
	doc := dom.Doc()
	var element dom.Element
	if c.ID == "" {
		element = doc.ByID("app")
	} else {
		element = dom.Doc().Query(fmt.Sprintf("[data-component-id='%s']", c.ID))
	}
	if element.IsNull() || element.IsUndefined() {
		return
	}

	selector := fmt.Sprintf(`[data-expr="%s"]`, exprID)
	nodes := element.Call("querySelectorAll", selector)
	for i := 0; i < nodes.Length(); i++ {
		node := nodes.Index(i)
		node.Set("innerHTML", fmt.Sprintf("%v", newValue))
	}
}

func updateClassExprBindings(c *HTMLComponent, exprID string, newValue any) {
	doc := dom.Doc()
	var element dom.Element
	if c.ID == "" {
		element = doc.ByID("app")
	} else {
		element = dom.Doc().Query(fmt.Sprintf("[data-component-id='%s']", c.ID))
	}
	if element.IsNull() || element.IsUndefined() {
		return
	}

	selector := fmt.Sprintf(`[data-expr-class="%s"]`, exprID)
	nodes := element.Call("querySelectorAll", selector)
	for i := 0; i < nodes.Length(); i++ {
		node := nodes.Index(i)
		newClassVal := strings.TrimSpace(fmt.Sprintf("%v", newValue))
		oldClassVal, ok := c.classExprContents[exprID]
		if !ok {
			return
		}
		c.classExprContents[exprID] = newClassVal

		currentClass := node.Get("className").String()
		if currentClass == "" {
			node.Set("className", newClassVal)
		} else {
			replaced := strings.Replace(currentClass, oldClassVal, newClassVal, 1)
			if replaced == currentClass && oldClassVal != "" && newClassVal != "" {
				replaced = currentClass + " " + newClassVal
			}
			node.Set("className", replaced)
		}
	}
}

func replacePropPlaceholders(template string, c *HTMLComponent) string {
	propRegex := reProp
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
	varRegex := rePluginVar
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
	cmdRegex := rePluginCmd
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
	varRegex := reHelperVar
	template = varRegex.ReplaceAllStringFunc(template, func(match string) string {
		name := varRegex.FindStringSubmatch(match)[1]
		c.hostVars = append(c.hostVars, name)

		expectedVal := ""
		if c.Props != nil {
			if v, ok := c.Props[name]; ok {
				expectedVal = fmt.Sprintf("%v", v)
			} else if v, ok := c.Props["h:"+name]; ok {
				expectedVal = fmt.Sprintf("%v", v)
			}
		}

		hash := sha1.Sum([]byte(expectedVal))
		expectedAttr := fmt.Sprintf("sha1:%x", hash)

		return fmt.Sprintf(`<span data-host-var="%s" data-host-expected="%s">%s</span>`,
			name, expectedAttr, html.EscapeString(expectedVal))
	})
	cmdRegex := reHelperCmd
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
	eventRegex := reEvent
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
	re := reRtIs
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
	expr := condition
	expr = strings.TrimPrefix(expr, "@if:")
	expr = strings.TrimPrefix(expr, "@else-if:")
	expr = strings.TrimSpace(expr)

	dependencies := extractDependencies(expr)

	lookup := func(name string) (any, bool) {
		if strings.HasPrefix(name, "store:") {
			parts := strings.Split(strings.TrimPrefix(name, "store:"), ".")
			if len(parts) == 3 {
				store := state.GlobalStoreManager.GetStore(parts[0], parts[1])
				if store != nil {
					return store.Get(parts[2]), true
				}
			}
			return nil, false
		}
		if strings.HasPrefix(name, "signal:") {
			sigName := strings.TrimPrefix(name, "signal:")
			if prop, ok := c.Props[sigName]; ok {
				if sig, ok := prop.(interface{ Read() any }); ok {
					return sig.Read(), true
				}
			}
			return nil, false
		}
		if strings.HasPrefix(name, "prop:") {
			propName := strings.TrimPrefix(name, "prop:")
			if v, ok := c.Props[propName]; ok {
				return v, true
			}
			return nil, false
		}
		if v, ok := c.Props[name]; ok {
			if sig, ok := v.(interface{ Read() any }); ok {
				return sig.Read(), true
			}
			return v, true
		}
		return nil, false
	}

	result, err := rtmleval.Bool(expr, lookup)
	if err != nil {
		Log().Debug("Condition evaluation error: %v", err)
		return false, dependencies
	}
	return result, dependencies
}

func extractDependencies(expr string) []ConditionDependency {
	var deps []ConditionDependency
	fields := depRegex.FindAllString(expr, -1)
	for _, f := range fields {
		if strings.HasPrefix(f, "store:") {
			parts := strings.Split(strings.TrimPrefix(f, "store:"), ".")
			if len(parts) == 3 {
				deps = append(deps, ConditionDependency{module: parts[0], storeName: parts[1], key: parts[2]})
			}
		} else if strings.HasPrefix(f, "signal:") {
			deps = append(deps, ConditionDependency{signal: strings.TrimPrefix(f, "signal:")})
		}
	}
	return deps
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
	tagRegex := reTagName
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
	re := reConditionalAttr
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
	forRegex := reFor
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
	foreachRegex := reForeach
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
			fieldRegex := regexp.MustCompile(fmt.Sprintf(`@%s\\.(\\w+(?:\\.\\w+)*)`, alias))
			iter = fieldRegex.ReplaceAllStringFunc(iter, func(fieldMatch string) string {
				fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
				if len(fieldParts) == 2 {
					if fieldValue, ok := resolveNestedKey(itemMap, fieldParts[1]); ok {
						return fmt.Sprintf("%v", fieldValue)
					}
				}
				return fieldMatch
			})
			propFieldRegex := regexp.MustCompile(fmt.Sprintf(`@prop:%s\\.(\\w+(?:\\.\\w+)*)`, alias))
			iter = propFieldRegex.ReplaceAllStringFunc(iter, func(fieldMatch string) string {
				fieldParts := propFieldRegex.FindStringSubmatch(fieldMatch)
				if len(fieldParts) == 2 {
					if fieldValue, ok := resolveNestedKey(itemMap, fieldParts[1]); ok {
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
			fieldRegex := regexp.MustCompile(fmt.Sprintf(`@%s\\.(\\w+(?:\\.\\w+)*)`, alias))
			iter = fieldRegex.ReplaceAllStringFunc(iter, func(fieldMatch string) string {
				fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
				if len(fieldParts) == 2 {
					if fieldValue, ok := resolveNestedKey(vMap, fieldParts[1]); ok {
						return fmt.Sprintf("%v", fieldValue)
					}
				}
				return fieldMatch
			})
			propFieldRegex := regexp.MustCompile(fmt.Sprintf(`@prop:%s\\.(\\w+(?:\\.\\w+)*)`, alias))
			iter = propFieldRegex.ReplaceAllStringFunc(iter, func(fieldMatch string) string {
				fieldParts := propFieldRegex.FindStringSubmatch(fieldMatch)
				if len(fieldParts) == 2 {
					if fieldValue, ok := resolveNestedKey(vMap, fieldParts[1]); ok {
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
	dom.BindStoreInputsForComponent(c.ID, node)
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

	dom.BindStoreInputsForComponent(c.ID, node)
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
	expr := condition
	expr = strings.TrimPrefix(expr, "@if:")
	expr = strings.TrimPrefix(expr, "@else-if:")
	return extractDependencies(expr), nil
}
