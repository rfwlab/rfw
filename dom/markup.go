package dom

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// only a whole, valueless attribute can be an event directive.
var reExpandEvent = regexp.MustCompile(`^@(on:)?(\w+(?:\.\w+)*):(\w+)$`)

// ExpandEvents rewrites @on:event:handler and @event:handler attributes into
// data-on-* attributes. Literal text, attribute values and raw-text elements
// are preserved. Templates go through it automatically; dynamic markup can use
// the same syntax before SetHTML.
func ExpandEvents(markup string) string {
	if !strings.Contains(markup, "@") {
		return markup
	}
	var sb strings.Builder
	last := 0
	// token boundaries keep comments and raw-text states out of the attribute
	// scan. Keep original bytes; serializing tokens would normalize user data.
	z := html.NewTokenizer(strings.NewReader(markup))
	offset := 0
	for {
		kind := z.Next()
		if kind == html.ErrorToken {
			break
		}
		size := len(z.Raw())
		if kind == html.StartTagToken || kind == html.SelfClosingTagToken {
			attrs := eventTagAttributes(markup[offset : offset+size])
			for _, attr := range attrs {
				attr[0] += offset
				attr[1] += offset
				name := markup[attr[0]:attr[1]]
				parts := reExpandEvent.FindStringSubmatchIndex(name)
				if parts == nil {
					continue
				}
				event, modifiers := name[parts[4]:parts[5]], ""
				if dot := strings.IndexByte(event, '.'); dot >= 0 {
					modifiers = strings.ReplaceAll(event[dot+1:], ".", ",")
					event = event[:dot]
				}
				if last == 0 {
					sb.Grow(len(markup))
				}
				sb.WriteString(markup[last:attr[0]])
				fmt.Fprintf(&sb, "data-on-%s=%q", event, name[parts[6]:parts[7]])
				if modifiers != "" {
					fmt.Fprintf(&sb, " data-on-%s-modifiers=%q", event, modifiers)
				}
				last = attr[1]
			}
		}
		offset += size
	}
	if last == 0 {
		return markup
	}
	sb.WriteString(markup[last:])
	return sb.String()
}

func eventTagSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}

// scan values as well as names: matching a directive-shaped substring in a
// URL or quoted value would turn data into event attributes.
func eventTagAttributes(markup string) [][2]int {
	i := 1
	for i < len(markup) && !eventTagSpace(markup[i]) && markup[i] != '/' && markup[i] != '>' {
		i++
	}
	var attrs [][2]int
	for i < len(markup) {
		for i < len(markup) && eventTagSpace(markup[i]) {
			i++
		}
		if i == len(markup) {
			break
		}
		if markup[i] == '>' {
			return attrs
		}
		if markup[i] == '/' {
			i++
			continue
		}
		start := i
		for i < len(markup) && !eventTagSpace(markup[i]) && markup[i] != '=' && markup[i] != '>' && markup[i] != '/' {
			i++
		}
		end := i
		for i < len(markup) && eventTagSpace(markup[i]) {
			i++
		}
		if i < len(markup) && markup[i] == '=' {
			i++
			for i < len(markup) && eventTagSpace(markup[i]) {
				i++
			}
			if i < len(markup) && (markup[i] == '\'' || markup[i] == '"') {
				quote := markup[i]
				i++
				for i < len(markup) && markup[i] != quote {
					i++
				}
				if i < len(markup) {
					i++
				}
			} else {
				for i < len(markup) && !eventTagSpace(markup[i]) && markup[i] != '>' {
					i++
				}
			}
		} else if end > start && markup[start] == '@' {
			attrs = append(attrs, [2]int{start, end})
		}
		if i == start {
			i++
		}
	}
	// an unfinished tag has no attributes to rewrite.
	return nil
}
