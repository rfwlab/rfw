//go:build js && wasm

package core

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

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
		case []Component:
			tmp := make([]interface{}, len(col))
			for i, v := range col {
				tmp[i] = v
			}
			collection = tmp
		case []*HTMLComponent:
			tmp := make([]interface{}, len(col))
			for i, v := range col {
				tmp[i] = v
			}
			collection = tmp
		case map[string]Component:
			tmp := make(map[string]interface{}, len(col))
			for k, v := range col {
				tmp[k] = v
			}
			collection = tmp
		case map[string]*HTMLComponent:
			tmp := make(map[string]interface{}, len(col))
			for k, v := range col {
				tmp[k] = v
			}
			collection = tmp
		}

		switch col := collection.(type) {
		case []interface{}:
			var result strings.Builder
			alias := aliases[0]
			for idx, item := range col {
				iterContent := loopContent
				if comp, ok := item.(Component); ok {
					placeholder := fmt.Sprintf("for-%s-%d", alias, idx)
					c.AddDependency(placeholder, comp)
					iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", alias), fmt.Sprintf("@include:%s", placeholder))
				} else if itemMap, ok := item.(map[string]interface{}); ok {
					fieldRegex := regexp.MustCompile(fmt.Sprintf("@prop:%s\\.(\\w+)", alias))
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
			for idx, k := range keys {
				v := col[k]
				iterContent := strings.ReplaceAll(loopContent, fmt.Sprintf("@prop:%s", keyAlias), k)
				if len(aliases) > 1 {
					if vMap, ok := v.(map[string]interface{}); ok {
						fieldRegex := regexp.MustCompile(fmt.Sprintf("@prop:%s\\.(\\w+)", valAlias))
						iterContent = fieldRegex.ReplaceAllStringFunc(iterContent, func(fieldMatch string) string {
							fieldParts := fieldRegex.FindStringSubmatch(fieldMatch)
							if len(fieldParts) == 2 {
								if fieldValue, exists := vMap[fieldParts[1]]; exists {
									return fmt.Sprintf("%v", fieldValue)
								}
							}
							return fieldMatch
						})
					} else if comp, ok := v.(Component); ok {
						placeholder := fmt.Sprintf("for-%s-%d", valAlias, idx)
						c.AddDependency(placeholder, comp)
						iterContent = strings.ReplaceAll(iterContent, fmt.Sprintf("@prop:%s", valAlias), fmt.Sprintf("@include:%s", placeholder))
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
