//go:build js && wasm

package core

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/internal/rendertrace"
	"github.com/rfwlab/rfw/v2/state"
)

func resolveNestedKey(m map[string]any, key string) (any, bool) {
	if strings.IndexByte(key, '.') < 0 {
		val, ok := m[key]
		return val, ok
	}
	parts := strings.Split(key, ".")
	val := any(m)
	for _, part := range parts {
		sub, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		val, ok = sub[part]
		if !ok {
			return nil, false
		}
	}
	return val, true
}

// reFor is compiled once. It used to be built inside replaceForPlaceholders,
// which meant recompiling the program on every render pass.
var reFor = regexp.MustCompile(`@for:(\w+(?:,\w+)?)\s+in\s+(\S+)([\s\S]*?)@endfor`)

func replaceForPlaceholders(template string, c *HTMLComponent) string {
	forRegex := reFor
	// loop ids are positional, so a re-render hands every loop the id its rows
	// already carry in the DOM
	c.forSeq = 0
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

		loopID := fmt.Sprintf("for-%s-%d", c.ID, c.forSeq)
		c.forSeq++

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
					// Rendering is coalesced per component; the common DOM
					// reconciler then limits mutation to changed loop rows.
					unsubscribe := store.OnChange(key, func(any) {
						if loopInHiddenConditionalBranch(c, loopID) {
							return
						}
						c.requestRender(rendertrace.Cause{
							Kind: "store", Module: module, Store: storeName, Key: key,
						})
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
		case []Component:
			tmp := make([]any, len(col))
			for i, v := range col {
				tmp[i] = v
			}
			collection = tmp
		case []*HTMLComponent:
			tmp := make([]any, len(col))
			for i, v := range col {
				tmp[i] = v
			}
			collection = tmp
		case map[string]Component:
			tmp := make(map[string]any, len(col))
			for k, v := range col {
				tmp[k] = v
			}
			collection = tmp
		case map[string]*HTMLComponent:
			tmp := make(map[string]any, len(col))
			for k, v := range col {
				tmp[k] = v
			}
			collection = tmp
		}

		rows, ok := expandForRows(c, aliases, loopContent, collection, loopID)
		if !ok {
			return match
		}
		return forAnchor(loopID) + rows
	})
}

// loopInHiddenConditionalBranch prevents data arriving for an unmounted view
// from repainting the visible component. The condition's own subscription will
// request a render when that branch becomes visible, using the latest data.
func loopInHiddenConditionalBranch(c *HTMLComponent, loopID string) bool {
	if c == nil || !c.mounted {
		return true
	}
	marker := fmt.Sprintf(`data-for-anchor="%s"`, loopID)
	for conditionID, content := range c.conditionContents {
		containsLoop := false
		for _, branch := range content.Branches {
			if strings.Contains(branch.Content, marker) {
				containsLoop = true
				break
			}
		}
		if !containsLoop {
			continue
		}

		root := dom.ComponentRoot(c.ID)
		if root.IsNull() || root.IsUndefined() {
			return true
		}
		condition := root.Query(fmt.Sprintf(`[data-condition="%s"]`, conditionID))
		if condition.IsNull() || condition.IsUndefined() {
			return true
		}
		selected := condition.Attr("data-condition-branch")
		for index, branch := range content.Branches {
			if fmt.Sprintf("%d", index) == selected {
				return !strings.Contains(branch.Content, marker)
			}
		}
		return true
	}
	return false
}

// forAnchor marks where a loop's rows begin. A template element carries no box
// and no layout, so it sits inside a flex or grid container without disturbing
// it, and an empty list still leaves the patch somewhere to insert into.
func forAnchor(loopID string) string {
	return fmt.Sprintf(`<template data-for-anchor="%s"></template>`, loopID)
}

// insertRowMarkers stamps the row key and the loop id on the row's opening tag,
// so a patch finds exactly the nodes the loop owns without touching the markup
// inside them.
func insertRowMarkers(content string, key any, loopID string) string {
	loc := reTagName.FindStringSubmatchIndex(content)
	if loc == nil {
		return content
	}
	attrs := ""
	openEnd := strings.IndexByte(content, '>')
	if openEnd < 0 || (!strings.Contains(content[:openEnd], "[key ") && !strings.Contains(content[:openEnd], "data-key=")) {
		attrs = fmt.Sprintf(` data-key="%v"`, key)
	}
	if loopID != "" {
		attrs += fmt.Sprintf(` data-for="%s"`, loopID)
	}
	return content[:loc[1]] + attrs + content[loc[1]:]
}

// expandForRows renders the loop body once per item. It reports false when the
// collection is not a shape the loop understands.
func expandForRows(c *HTMLComponent, aliases []string, loopContent string, collection any, loopID string) (string, bool) {
	switch col := collection.(type) {
	case nil:
		// unset store key: no rows, and the anchor keeps the spot so the first
		// value that lands can be patched in
		return "", true
	case []any:
		var result strings.Builder
		alias := aliases[0]
		for idx, item := range col {
			iterContent := loopContent
			if comp, ok := item.(Component); ok {
				placeholder := fmt.Sprintf("for-%s-%d", alias, idx)
				c.AddDependency(placeholder, comp)
				iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", alias), fmt.Sprintf("@include:%s", placeholder))
			} else if itemMap, ok := item.(map[string]any); ok {
				iterContent = substituteItemFields(iterContent, alias, itemMap)
			} else {
				iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", alias), escapeValue(item))
			}
			iterContent = insertRowMarkers(iterContent, idx, loopID)
			result.WriteString(iterContent)
		}
		return result.String(), true
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
		for idx, k := range keys {
			v := col[k]
			iterContent := strings.ReplaceAll(loopContent, fmt.Sprintf("@prop:%s", keyAlias), escapeValue(k))
			if len(aliases) > 1 {
				if vMap, ok := v.(map[string]any); ok {
					iterContent = substituteItemFields(iterContent, valAlias, vMap)
				} else if comp, ok := v.(Component); ok {
					placeholder := fmt.Sprintf("for-%s-%d", valAlias, idx)
					c.AddDependency(placeholder, comp)
					iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", valAlias), fmt.Sprintf("@include:%s", placeholder))
				} else {
					iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@rawprop:%s", valAlias), fmt.Sprintf("%v", v))
					iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", valAlias), escapeValue(v))
				}
			}
			iterContent = insertRowMarkers(iterContent, k, loopID)
			result.WriteString(iterContent)
		}
		return result.String(), true
	default:
		return "", false
	}
}

// substituteItemFields fills the @prop / @rawprop field references of one item.
// aliasRe caches the per-alias @prop / @rawprop matchers. substituteItemFields
// runs once per item of a loop, so compiling these inline cost two regexp
// compilations per row: a hundred-row list paid two hundred of them on every
// render.
var aliasRe sync.Map // alias -> [2]*regexp.Regexp{raw, field}

func aliasRegexps(alias string) (*regexp.Regexp, *regexp.Regexp) {
	if v, ok := aliasRe.Load(alias); ok {
		p := v.([2]*regexp.Regexp)
		return p[0], p[1]
	}
	pair := [2]*regexp.Regexp{
		regexp.MustCompile(fmt.Sprintf(`@rawprop:%s\.(\w+(?:\.\w+)*)`, regexp.QuoteMeta(alias))),
		regexp.MustCompile(fmt.Sprintf(`@prop:%s\.(\w+(?:\.\w+)*)`, regexp.QuoteMeta(alias))),
	}
	aliasRe.Store(alias, pair)
	return pair[0], pair[1]
}

func substituteItemFields(content, alias string, item map[string]any) string {
	rawRegex, fieldRegex := aliasRegexps(alias)
	content = substituteFieldRefs(rawRegex, content, item, false)
	return substituteFieldRefs(fieldRegex, content, item, true)
}

// substituteFieldRefs replaces every match of re, whose first group is the
// field path, with the item's value. Going through the match indexes keeps the
// pattern to one execution per row: ReplaceAllStringFunc would run it a second
// time inside the callback only to reach the group.
func substituteFieldRefs(re *regexp.Regexp, content string, item map[string]any, escape bool) string {
	locs := re.FindAllStringSubmatchIndex(content, -1)
	if locs == nil {
		return content
	}
	var sb strings.Builder
	sb.Grow(len(content))
	last := 0
	for _, loc := range locs {
		fieldValue, ok := resolveNestedKey(item, content[loc[2]:loc[3]])
		if !ok {
			// an unresolved reference stays in the markup, as it did when the
			// callback returned the match untouched
			continue
		}
		sb.WriteString(content[last:loc[0]])
		if escape {
			sb.WriteString(escapeValue(fieldValue))
		} else if s, ok := fieldValue.(string); ok {
			sb.WriteString(s)
		} else {
			fmt.Fprintf(&sb, "%v", fieldValue)
		}
		last = loc[1]
	}
	sb.WriteString(content[last:])
	return sb.String()
}
