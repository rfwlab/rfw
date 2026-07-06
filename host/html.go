package host

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

const hostVarAttr = "data-host-var"
const hostExpectedAttr = "data-host-expected"
const expectationAlg = "sha1"

func encodeExpected(value string) string {
	sum := sha1.Sum([]byte(value))
	return fmt.Sprintf("%s:%s", expectationAlg, hex.EncodeToString(sum[:]))
}

func hostVarTag(tag, name string, value any) string {
	v := fmt.Sprintf("%v", value)
	return fmt.Sprintf(`<%s %s="%s" %s="%s">%s</%s>`,
		tag, hostVarAttr, name, hostExpectedAttr, encodeExpected(v), v, tag)
}

func Span(name string, value any) string {
	return hostVarTag("span", name, value)
}

func Div(name string, value any) string {
	return hostVarTag("div", name, value)
}

func P(name string, value any) string {
	return hostVarTag("p", name, value)
}

func Tag(tag, name string, value any) string {
	return hostVarTag(tag, name, value)
}

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