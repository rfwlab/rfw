//go:build js && wasm

package core

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

func resolveNestedKey(m map[string]any, key string) (any, bool) {
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

func replaceForPlaceholders(template string, c *HTMLComponent) string {
	forRegex := regexp.MustCompile(`@for:(\w+(?:,\w+)?)\s+in\s+(\S+)([\s\S]*?)@endfor`)
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
					// a list that changes should cost its own rows, not a
					// re-render of the whole component: patch the loop subtree
					// when the body allows it and fall back otherwise
					unsubscribe := store.OnChange(key, func(newValue any) {
						if patchForLoop(c, loopID, aliases, loopContent, newValue) {
							return
						}
						dom.UpdateMountedDOM(c.ID, c.RenderFresh())
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
	attrs := fmt.Sprintf(` data-key="%v"`, key)
	if loopID != "" {
		attrs += fmt.Sprintf(` data-for="%s"`, loopID)
	}
	return content[:loc[1]] + attrs + content[loc[1]:]
}

// singleRootRow reports whether the loop body renders exactly one element per
// row. Multi-root rows carry the loop id on their first element only, so the
// patch could not clean them up and falls back to a render instead.
func singleRootRow(body string) bool {
	depth, roots := 0, 0
	for _, m := range reAnyTag.FindAllStringSubmatch(body, -1) {
		closing, name, selfClosing := m[1] == "/", strings.ToLower(m[2]), strings.HasSuffix(m[0], "/>")
		if voidElements[name] || selfClosing {
			if depth == 0 {
				roots++
			}
			continue
		}
		if closing {
			depth--
			continue
		}
		if depth == 0 {
			roots++
		}
		depth++
	}
	return roots == 1
}

var reAnyTag = regexp.MustCompile(`<(/?)([a-zA-Z][\w-]*)[^>]*>`)

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"source": true, "track": true, "wbr": true,
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
func substituteItemFields(content, alias string, item map[string]any) string {
	rawRegex := regexp.MustCompile(fmt.Sprintf(`@rawprop:%s\.(\w+(?:\.\w+)*)`, alias))
	content = rawRegex.ReplaceAllStringFunc(content, func(fieldMatch string) string {
		fieldParts := rawRegex.FindStringSubmatch(fieldMatch)
		if len(fieldParts) == 2 {
			if fieldValue, ok := resolveNestedKey(item, fieldParts[1]); ok {
				return fmt.Sprintf("%v", fieldValue)
			}
		}
		return fieldMatch
	})
	fieldRegex := regexp.MustCompile(fmt.Sprintf(`@prop:%s\.(\w+(?:\.\w+)*)`, alias))
	return fieldRegex.ReplaceAllStringFunc(content, func(fieldMatch string) string {
		fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
		if len(fieldParts) == 2 {
			if fieldValue, ok := resolveNestedKey(item, fieldParts[1]); ok {
				return escapeValue(fieldValue)
			}
		}
		return fieldMatch
	})
}
