package host

import (
	"fmt"
	"html"
	"strings"
)

const hostVarAttr = "data-host-var"
const hostExpectedAttr = "data-host-expected"

func validTagName(tag string) bool {
	if tag == "" || !isASCIILetter(rune(tag[0])) {
		return false
	}
	for _, char := range tag {
		if isASCIILetter(char) || char >= '0' && char <= '9' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func isASCIILetter(char rune) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
}

// hostVarTag builds a host variable element. The value is HTML-escaped so
// user-derived data cannot inject markup through the initial snapshot.
func hostVarTag(tag, name string, value any, escape bool) string {
	if !validTagName(tag) {
		return ""
	}
	v := fmt.Sprintf("%v", value)
	body := v
	expected := ""
	if escape {
		body = html.EscapeString(v)
		expected = html.EscapeString(v)
	}
	return fmt.Sprintf(`<%s %s="%s" %s="%s">%s</%s>`,
		tag, hostVarAttr, html.EscapeString(name), hostExpectedAttr, expected, body, tag)
}

// Span renders an escaped host variable in a span.
func Span(name string, value any) string {
	return hostVarTag("span", name, value, true)
}

// Div renders an escaped host variable in a div.
func Div(name string, value any) string {
	return hostVarTag("div", name, value, true)
}

// P renders an escaped host variable in a paragraph.
func P(name string, value any) string {
	return hostVarTag("p", name, value, true)
}

// Tag renders an escaped host variable with tag.
func Tag(tag, name string, value any) string {
	return hostVarTag(tag, name, value, true)
}

// RawTag builds a host variable element without escaping the value. It is the
// explicit trust API for markup values: only pass HTML you generated or
// sanitized yourself, never user-derived data.
func RawTag(tag, name string, value any) string {
	return hostVarTag(tag, name, value, false)
}

// Raw marks a fragment as trusted HTML and returns it unchanged. It exists to
// make raw injection points explicit at call sites: anything passed through
// Raw ends up in the client DOM unescaped via InitSnapshot.HTML.
func Raw(html string) string {
	return html
}

// Join concatenates rendered host fragments.
func Join(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}
