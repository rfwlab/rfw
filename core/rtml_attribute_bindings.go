//go:build js && wasm

package core

import (
	"fmt"
	"strings"
)

// Read bindings (@store, @rawstore and @signal without the :w suffix) expand
// into a <span> holding the value: they bind text and markup. Written inside a
// start tag that span closes the attribute early, so the browser gets a
// half-built element and the tail of the tag lands in the document as text.
// The renderer refuses such a template instead of emitting the corruption.
// The writable :w form is untouched: it survives the render as a placeholder
// that the DOM layer binds on form attributes.
func validateAttributeBindings(template string) error {
	for i := 0; i < len(template); {
		if template[i] != '<' {
			i++
			continue
		}
		rest := template[i:]
		switch {
		case strings.HasPrefix(rest, "<!--"):
			end := strings.Index(rest, "-->")
			if end < 0 {
				return nil
			}
			i += end + len("-->")
		case strings.HasPrefix(rest, "</"), strings.HasPrefix(rest, "<!"), strings.HasPrefix(rest, "<?"):
			// closing tags, doctypes and processing instructions have no attributes
			end := strings.IndexByte(rest, '>')
			if end < 0 {
				return nil
			}
			i += end + 1
		case len(rest) > 1 && isTagNameStart(rest[1]):
			next, tag, err := scanStartTagBindings(template, i)
			if err != nil {
				return err
			}
			i = next
			if isRawTextTag(tag) {
				// inside script/style a '<' is text, not a tag start
				end := indexClosingTag(template[i:], tag)
				if end < 0 {
					return nil
				}
				i += end
			}
		default:
			i++
		}
	}
	return nil
}

// scanStartTagBindings walks the attributes of the start tag opening at start
// and returns the offset just past it together with the tag name.
func scanStartTagBindings(template string, start int) (int, string, error) {
	i := start + 1
	nameStart := i
	for i < len(template) && isTagNameByte(template[i]) {
		i++
	}
	tag := template[nameStart:i]

	for i < len(template) {
		if template[i] == '>' {
			return i + 1, tag, nil
		}
		if isTemplateSpace(template[i]) || template[i] == '/' {
			i++
			continue
		}

		nameStart = i
		for i < len(template) && !isTemplateSpace(template[i]) &&
			template[i] != '=' && template[i] != '>' && template[i] != '/' {
			i++
		}
		attr := template[nameStart:i]

		// HTML allows whitespace on either side of the '=', so the value can sit
		// past the end of the name token: peek before calling the attribute
		// valueless, otherwise the value is scanned as a nameless attribute and
		// the report cannot name the attribute that holds the binding.
		if eq := skipTemplateSpace(template, i); eq < len(template) && template[eq] == '=' {
			i = skipTemplateSpace(template, eq+1)
			var value string
			if i < len(template) && (template[i] == '"' || template[i] == '\'') {
				quote := template[i]
				i++
				valueStart := i
				for i < len(template) && template[i] != quote {
					i++
				}
				value = template[valueStart:i]
				if i < len(template) {
					i++
				}
			} else {
				valueStart := i
				for i < len(template) && !isTemplateSpace(template[i]) && template[i] != '>' {
					i++
				}
				value = template[valueStart:i]
			}
			if binding, ok := findReadBinding(value); ok {
				return i, tag, attributeBindingError(binding, attr, tag)
			}
			continue
		}

		// a valueless attribute that is itself a binding corrupts the tag too
		if binding, ok := findReadBinding(attr); ok {
			return i, tag, attributeBindingError(binding, attr, tag)
		}
	}
	return i, tag, nil
}

// findReadBinding reports the first read-only binding in s, reusing the
// renderer's own patterns so the two never disagree on what a binding is.
func findReadBinding(s string) (string, bool) {
	if !strings.Contains(s, "@") {
		return "", false
	}
	if match := reRawStore.FindString(s); match != "" {
		return match, true
	}
	for _, parts := range reStore.FindAllStringSubmatch(s, -1) {
		if parts[4] != ":w" {
			return parts[0], true
		}
	}
	for _, parts := range reSignal.FindAllStringSubmatch(s, -1) {
		if parts[2] != ":w" {
			return parts[0], true
		}
	}
	return "", false
}

func attributeBindingError(binding, attr, tag string) error {
	return fmt.Errorf("read-only binding %s cannot be used in attribute %q of <%s>: "+
		"@store, @rawstore and @signal render content, not attribute values; "+
		"use @if to control whether the element is rendered, or the :w form for two-way form bindings",
		binding, attr, tag)
}

// isRawTextTag reports whether the element holds text rather than markup, so
// its content must not be scanned for tags.
func isRawTextTag(tag string) bool {
	switch strings.ToLower(tag) {
	case "script", "style", "textarea", "title":
		return true
	}
	return false
}

func indexClosingTag(s, tag string) int {
	return strings.Index(strings.ToLower(s), "</"+strings.ToLower(tag))
}

func isTagNameStart(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func isTagNameByte(b byte) bool {
	return isTagNameStart(b) || b >= '0' && b <= '9' || b == '-' || b == '_'
}

func skipTemplateSpace(s string, i int) int {
	for i < len(s) && isTemplateSpace(s[i]) {
		i++
	}
	return i
}

func isTemplateSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	}
	return false
}
