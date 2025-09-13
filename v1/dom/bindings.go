//go:build js && wasm

package dom

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// binding represents a precompiled event binding.
type binding struct {
	Path      []int
	Event     string
	Handler   string
	Modifiers []string
}

var (
	// precompiledByName caches bindings per component name so they are
	// generated only once per run.
	precompiledByName = make(map[string][]binding)
	compiledBindings  = make(map[string][]binding)
)

// RegisterBindings generates and associates bindings for a component instance.
func RegisterBindings(id, name, template string) {
	if bs, ok := precompiledByName[name]; ok {
		compiledBindings[id] = bs
		return
	}
	bs, err := parseTemplate(template)
	if err != nil {
		return
	}
	precompiledByName[name] = bs
	compiledBindings[id] = bs
}

func parseTemplate(tpl string) ([]binding, error) {
	processed := replaceEventHandlers(tpl)
	node, err := html.Parse(strings.NewReader(processed))
	if err != nil {
		return nil, err
	}
	return collectBindings(node, nil), nil
}

func collectBindings(n *html.Node, path []int) []binding {
	var res []binding
	if n.Type == html.ElementNode {
		attrs := map[string]string{}
		for _, a := range n.Attr {
			attrs[a.Key] = a.Val
		}
		for k, v := range attrs {
			if strings.HasPrefix(k, "data-on-") && !strings.HasSuffix(k, "-modifiers") {
				event := strings.TrimPrefix(k, "data-on-")
				mods := []string{}
				if m, ok := attrs[fmt.Sprintf("data-on-%s-modifiers", event)]; ok && m != "" {
					for _, s := range strings.Split(m, ",") {
						s = strings.TrimSpace(s)
						if s != "" {
							mods = append(mods, s)
						}
					}
				}
				res = append(res, binding{Path: append([]int(nil), path...), Event: event, Handler: v, Modifiers: mods})
			}
		}
	}
	child := n.FirstChild
	idx := 0
	for child != nil {
		res = append(res, collectBindings(child, append(path, idx))...)
		child = child.NextSibling
		idx++
	}
	return res
}

// eventRegex matches event directives (e.g. @on:click:handler) ensuring the
// handler is terminated by whitespace, a self-closing slash or the end of the
// tag. The terminating character is captured to preserve it during replacement
// and avoid matching constructs like store bindings within attribute values.
var eventRegex = regexp.MustCompile(`@(on:)?(\w+(?:\.\w+)*):(\w+)([\s>/])`)

func replaceEventHandlers(template string) string {
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
