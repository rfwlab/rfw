package dom

import (
	"strings"

	"golang.org/x/net/html"
)

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
	var attrs [][2]int
	offset := 0
	for {
		kind := z.Next()
		if kind == html.ErrorToken {
			break
		}
		size := len(z.Raw())
		if kind == html.StartTagToken || kind == html.SelfClosingTagToken {
			attrs = eventTagAttributes(markup[offset:offset+size], attrs[:0])
			for _, attr := range attrs {
				attr[0] += offset
				attr[1] += offset
				name := markup[attr[0]:attr[1]]
				event, handler := eventDirective(name)
				if event == "" {
					continue
				}
				modifiers := ""
				if dot := strings.IndexByte(event, '.'); dot >= 0 {
					modifiers = strings.ReplaceAll(event[dot+1:], ".", ",")
					event = event[:dot]
				}
				if last == 0 {
					sb.Grow(len(markup))
				}
				sb.WriteString(markup[last:attr[0]])
				sb.WriteString("data-on-")
				sb.WriteString(event)
				sb.WriteString("=\"")
				sb.WriteString(handler)
				sb.WriteByte('"')
				if modifiers != "" {
					sb.WriteString(" data-on-")
					sb.WriteString(event)
					sb.WriteString("-modifiers=\"")
					sb.WriteString(modifiers)
					sb.WriteByte('"')
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
func eventTagAttributes(markup string, attrs [][2]int) [][2]int {
	i := 1
	for i < len(markup) && !eventTagSpace(markup[i]) && markup[i] != '/' && markup[i] != '>' {
		i++
	}
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

func eventDirective(name string) (event, handler string) {
	if len(name) < 4 || name[0] != '@' {
		return "", ""
	}
	colon := strings.LastIndexByte(name, ':')
	if colon < 2 || colon == len(name)-1 {
		return "", ""
	}
	event, handler = name[1:colon], name[colon+1:]
	event = strings.TrimPrefix(event, "on:")
	if !eventWord(event, true) || !eventWord(handler, false) {
		return "", ""
	}
	return event, handler
}

// directive words stay ASCII, as in the original event syntax. Validated words
// need no quoting escapes when written into HTML attributes.
func eventWord(word string, dots bool) bool {
	if word == "" {
		return false
	}
	for i := 0; i < len(word); i++ {
		c := word[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' {
			continue
		}
		if dots && c == '.' && i > 0 && i < len(word)-1 && word[i-1] != '.' {
			continue
		}
		return false
	}
	return true
}
