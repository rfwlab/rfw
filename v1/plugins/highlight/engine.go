package highlight

import (
	"html"
	"regexp"
	"strings"
)

var (
	goKeywords = []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}
	goKeywordRe = regexp.MustCompile(`\b(` + strings.Join(goKeywords, "|") + `)\b`)
	goStringRe  = regexp.MustCompile(`&#34;[^\n]*?&#34;`)
	goCommentRe = regexp.MustCompile(`(?m)//.*?$|/\*[\s\S]*?\*/`)

	rtmlTagRe        = regexp.MustCompile(`&lt;/?[\s\S]*?&gt;`)
	rtmlStringRe     = regexp.MustCompile(`&#34;.*?&#34;`)
	rtmlAttrRe       = regexp.MustCompile(`\s+([a-zA-Z_:][\w:.-]*)=`)
	rtmlCmdAttrRe    = regexp.MustCompile(`\s+(@[\w:.-]+)`)
	rtmlCmdRe        = regexp.MustCompile(`@[\w:.-]+`)
	rtmlStandaloneRe = regexp.MustCompile(`(?m)^(\s*)(@[\w:.-]+)`)
	rtmlInterpRe     = regexp.MustCompile(`\{[^\}]+\}`)
)

// Highlight returns HTML with basic syntax highlighting for supported languages.
func Highlight(code, lang string) (string, bool) {
	switch lang {
	case "go":
		return highlightGo(code), true
	case "rtml":
		return highlightRTML(code), true
	default:
		return "", false
	}
}

func highlightGo(code string) string {
	esc := html.EscapeString(code)
	esc = goCommentRe.ReplaceAllStringFunc(esc, func(m string) string {
		return `<span class="hl-comment">` + m + `</span>`
	})
	esc = goStringRe.ReplaceAllStringFunc(esc, func(m string) string {
		return `<span class="hl-string">` + m + `</span>`
	})
	esc = goKeywordRe.ReplaceAllString(esc, `<span class="hl-kw">$1</span>`)
	return esc
}

func highlightRTML(code string) string {
	esc := html.EscapeString(code)
	esc = rtmlTagRe.ReplaceAllStringFunc(esc, func(m string) string {
		s := rtmlAttrRe.ReplaceAllString(m, ` <span class="hl-attr">$1</span>=`)
		s = rtmlCmdAttrRe.ReplaceAllString(s, ` <span class="hl-cmd">$1</span>`)
		s = rtmlStringRe.ReplaceAllStringFunc(s, func(str string) string {
			quote := "&#34;"
			inner := str[len(quote) : len(str)-len(quote)]
			inner = rtmlCmdRe.ReplaceAllStringFunc(inner, func(cmd string) string {
				return `<span class="hl-cmd">` + cmd + `</span>`
			})
			return `<span class="hl-string">` + quote + inner + quote + `</span>`
		})
		return `<span class="hl-tag">` + s + `</span>`
	})
	esc = rtmlStandaloneRe.ReplaceAllString(esc, `$1<span class="hl-cmd">$2</span>`)
	esc = rtmlInterpRe.ReplaceAllStringFunc(esc, func(m string) string {
		return `<span class="hl-var">` + m + `</span>`
	})
	return esc
}
