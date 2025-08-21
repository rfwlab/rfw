package rtml

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/rfwlab/rfw/v1/state"
)

// Dependency represents a renderable component used for includes.
type Dependency interface {
	Render() string
}

// Context carries rendering data for templates.
type Context struct {
	Props        map[string]any
	Slots        map[string]any
	Dependencies map[string]Dependency
}

// Replace runs a minimal RTML rendering pipeline for server-side rendering.
func Replace(template string, ctx Context) string {
	rendered := replacePropPlaceholders(template, ctx)
	rendered = replaceIncludePlaceholders(ctx, rendered)
	rendered = replaceSlotPlaceholders(rendered, ctx)
	rendered = replaceForPlaceholders(rendered, ctx)
	rendered = replaceConditionals(rendered, ctx)
	rendered = replaceStorePlaceholders(rendered, ctx)
	rendered = replaceRtIsAttributes(rendered, ctx)
	rendered = replaceIncludePlaceholders(ctx, rendered)
	return rendered
}

func replacePropPlaceholders(template string, ctx Context) string {
	if ctx.Props == nil {
		return template
	}
	for k, v := range ctx.Props {
		pattern := fmt.Sprintf(`{{\s*%s\s*}}`, regexp.QuoteMeta(k))
		re := regexp.MustCompile(pattern)
		template = re.ReplaceAllString(template, fmt.Sprintf("%v", v))
	}
	return template
}

func replaceSlotPlaceholders(template string, ctx Context) string {
	if ctx.Slots == nil {
		return template
	}
	for name, content := range ctx.Slots {
		placeholder := fmt.Sprintf("@slot:%s", name)
		template = strings.ReplaceAll(template, placeholder, fmt.Sprintf("%v", content))
	}
	return template
}

func replaceIncludePlaceholders(ctx Context, template string) string {
	if ctx.Dependencies == nil {
		return template
	}
	for name, dep := range ctx.Dependencies {
		include := fmt.Sprintf("@include:%s", name)
		template = strings.ReplaceAll(template, include, dep.Render())
	}
	return template
}

func replaceForPlaceholders(template string, ctx Context) string {
	forRegex := regexp.MustCompile(`@for:(\w+(?:,\w+)?)\s+in\s+(\S+)([\s\S]*?)@endfor`)
	return forRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := forRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		aliasesPart := parts[1]
		expr := parts[2]
		loopContent := parts[3]

		aliases := strings.Split(aliasesPart, ",")
		for i := range aliases {
			aliases[i] = strings.TrimSpace(aliases[i])
		}

		if strings.Contains(expr, "..") {
			rangeParts := strings.Split(expr, "..")
			if len(rangeParts) != 2 {
				return match
			}
			start, ok1 := resolveNumber(rangeParts[0], ctx)
			end, ok2 := resolveNumber(rangeParts[1], ctx)
			if !ok1 || !ok2 {
				return match
			}
			var sb strings.Builder
			for i := start; i <= end; i++ {
				iter := strings.ReplaceAll(loopContent, fmt.Sprintf("@prop:%s", aliases[0]), fmt.Sprintf("%d", i))
				sb.WriteString(iter)
			}
			return sb.String()
		}

		collection, ok := ctx.Props[expr]
		if !ok {
			if strings.HasPrefix(expr, "store:") {
				parts := strings.Split(strings.TrimPrefix(expr, "store:"), ".")
				if len(parts) == 3 {
					if store := state.GlobalStoreManager.GetStore(parts[0], parts[1]); store != nil {
						collection = store.Get(parts[2])
						ok = true
					}
				}
			}
			if !ok {
				return match
			}
		}

		val := reflect.ValueOf(collection)
		switch val.Kind() {
		case reflect.Slice, reflect.Array:
			var sb strings.Builder
			alias := aliases[0]
			for i := 0; i < val.Len(); i++ {
				item := val.Index(i).Interface()
				iter := loopContent
				iter = replaceAlias(iter, alias, item)
				sb.WriteString(iter)
			}
			return sb.String()
		case reflect.Map:
			keys := val.MapKeys()
			aliasKey := aliases[0]
			aliasVal := aliasKey
			if len(aliases) > 1 {
				aliasVal = aliases[1]
			}
			var sb strings.Builder
			for _, k := range keys {
				v := val.MapIndex(k).Interface()
				iter := loopContent
				iter = strings.ReplaceAll(iter, fmt.Sprintf("@prop:%s", aliasKey), fmt.Sprintf("%v", k.Interface()))
				iter = replaceAlias(iter, aliasVal, v)
				sb.WriteString(iter)
			}
			return sb.String()
		default:
			return match
		}
	})
}

func replaceAlias(template, alias string, val any) string {
	switch v := val.(type) {
	case Dependency:
		return strings.ReplaceAll(template, fmt.Sprintf("@prop:%s", alias), v.Render())
	case map[string]any:
		re := regexp.MustCompile(fmt.Sprintf(`@prop:%s\\.(\\w+)`, regexp.QuoteMeta(alias)))
		return re.ReplaceAllStringFunc(template, func(m string) string {
			parts := re.FindStringSubmatch(m)
			if len(parts) == 2 {
				if fv, ok := v[parts[1]]; ok {
					return fmt.Sprintf("%v", fv)
				}
			}
			return m
		})
	default:
		return strings.ReplaceAll(template, fmt.Sprintf("@prop:%s", alias), fmt.Sprintf("%v", v))
	}
}

func resolveNumber(expr string, ctx Context) (int, bool) {
	expr = strings.TrimSpace(expr)
	if v, ok := ctx.Props[expr]; ok {
		switch n := v.(type) {
		case int:
			return n, true
		case int64:
			return int(n), true
		case float64:
			return int(n), true
		case string:
			i, err := strconv.Atoi(n)
			if err == nil {
				return i, true
			}
		}
	}
	if strings.HasPrefix(expr, "store:") {
		parts := strings.Split(strings.TrimPrefix(expr, "store:"), ".")
		if len(parts) == 3 {
			if store := state.GlobalStoreManager.GetStore(parts[0], parts[1]); store != nil {
				if v := store.Get(parts[2]); v != nil {
					switch n := v.(type) {
					case int:
						return n, true
					case int64:
						return int(n), true
					case float64:
						return int(n), true
					case string:
						i, err := strconv.Atoi(n)
						if err == nil {
							return i, true
						}
					}
				}
			}
		}
	}
	i, err := strconv.Atoi(expr)
	if err == nil {
		return i, true
	}
	return 0, false
}

// --- Conditional rendering ---

type node interface {
	Render(ctx Context) string
}

type textNode struct{ Text string }

func (t *textNode) Render(ctx Context) string { return t.Text }

type conditionalBranch struct {
	Condition string
	Nodes     []node
}

type conditionalNode struct{ Branches []conditionalBranch }

func (cn *conditionalNode) Render(ctx Context) string {
	for _, br := range cn.Branches {
		if br.Condition == "" || evaluateCondition(br.Condition, ctx) {
			var sb strings.Builder
			for _, n := range br.Nodes {
				sb.WriteString(n.Render(ctx))
			}
			return sb.String()
		}
	}
	return ""
}

func replaceConditionals(template string, ctx Context) string {
	if !strings.Contains(template, "@if:") {
		return template
	}
	nodes, err := parseTemplate(template)
	if err != nil {
		return template
	}
	var sb strings.Builder
	for _, n := range nodes {
		sb.WriteString(n.Render(ctx))
	}
	return sb.String()
}

func parseTemplate(template string) ([]node, error) {
	lines := strings.Split(template, "\n")
	idx := 0
	return parseBlock(lines, &idx)
}

func parseBlock(lines []string, idx *int) ([]node, error) {
	var nodes []node
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
			nodes = append(nodes, &textNode{Text: line + "\n"})
			*idx++
		}
	}
	return nodes, nil
}

func parseConditional(lines []string, idx *int, firstCond string) (node, error) {
	n := &conditionalNode{}
	children, err := parseBlock(lines, idx)
	if err != nil {
		return nil, err
	}
	n.Branches = append(n.Branches, conditionalBranch{Condition: firstCond, Nodes: children})

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
			n.Branches = append(n.Branches, conditionalBranch{Condition: cond, Nodes: children})
		case trimmed == "@else":
			*idx++
			children, err := parseBlock(lines, idx)
			if err != nil {
				return nil, err
			}
			n.Branches = append(n.Branches, conditionalBranch{Condition: "", Nodes: children})
		case trimmed == "@endif":
			*idx++
			return n, nil
		default:
			*idx++
		}
	}
	return n, nil
}

func evaluateCondition(condition string, ctx Context) bool {
	parts := strings.Split(condition, "==")
	if len(parts) != 2 {
		return false
	}
	left := strings.TrimSpace(parts[0])
	right := strings.Trim(parts[1], "\"' ")
	left = strings.TrimPrefix(left, "@if:")
	left = strings.TrimPrefix(left, "@else-if:")
	if strings.HasPrefix(left, "prop:") {
		key := strings.TrimPrefix(left, "prop:")
		if v, ok := ctx.Props[key]; ok {
			return fmt.Sprintf("%v", v) == right
		}
	} else if strings.HasPrefix(left, "store:") {
		parts := strings.Split(strings.TrimPrefix(left, "store:"), ".")
		if len(parts) == 3 {
			if store := state.GlobalStoreManager.GetStore(parts[0], parts[1]); store != nil {
				if v := store.Get(parts[2]); v != nil {
					return fmt.Sprintf("%v", v) == right
				}
			}
		}
	}
	return false
}

// --- Stores ---

func replaceStorePlaceholders(template string, ctx Context) string {
	re := regexp.MustCompile(`@store:(\w+)\.(\w+)\.(\w+)(:w)?`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 5 {
			return match
		}
		if parts[4] == ":w" {
			return match
		}
		if store := state.GlobalStoreManager.GetStore(parts[1], parts[2]); store != nil {
			if v := store.Get(parts[3]); v != nil {
				return fmt.Sprintf("%v", v)
			}
		}
		return match
	})
}

// --- rt-is ---

func replaceRtIsAttributes(template string, ctx Context) string {
	re := regexp.MustCompile(`<([a-zA-Z0-9]+)([^>]*)rt-is="([^"]+)"[^>]*/?>`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		name := parts[3]
		if dep, ok := ctx.Dependencies[name]; ok {
			return dep.Render()
		}
		return match
	})
}
