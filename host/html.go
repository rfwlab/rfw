package host

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"html"
	"strings"
)

const hostVarAttr = "data-host-var"
const hostExpectedAttr = "data-host-expected"
const expectationAlg = "sha1"

func encodeExpected(value string) string {
	sum := sha1.Sum([]byte(value))
	return fmt.Sprintf("%s:%s", expectationAlg, hex.EncodeToString(sum[:]))
}

// hostVarTag builds a host variable element. The value is HTML-escaped so
// user-derived data cannot inject markup through the initial snapshot; the
// expectation hash is computed over the unescaped value because the client
// hashes the element's text content, which the browser unescapes.
func hostVarTag(tag, name string, value any, escape bool) string {
	v := fmt.Sprintf("%v", value)
	body := v
	if escape {
		body = html.EscapeString(v)
	}
	return fmt.Sprintf(`<%s %s="%s" %s="%s">%s</%s>`,
		tag, hostVarAttr, name, hostExpectedAttr, encodeExpected(v), body, tag)
}

func Span(name string, value any) string {
	return hostVarTag("span", name, value, true)
}

func Div(name string, value any) string {
	return hostVarTag("div", name, value, true)
}

func P(name string, value any) string {
	return hostVarTag("p", name, value, true)
}

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

func Join(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}
