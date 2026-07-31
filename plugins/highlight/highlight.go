//go:build js && wasm

// Package highlight adds syntax highlighting to code blocks.
package highlight

import (
	"encoding/json"
	"strings"

	"github.com/rfwlab/rfw/v2/core"
	dom "github.com/rfwlab/rfw/v2/dom"
	js "github.com/rfwlab/rfw/v2/js"
)

// Plugin installs syntax highlighting helpers.
type Plugin struct{}

// New creates a syntax highlighting plugin.
func New() *Plugin { return &Plugin{} }

// Build accepts the plugin build configuration.
func (p *Plugin) Build(json.RawMessage) error { return nil }

// All finds all <pre><code> blocks in the document and replaces
// their contents with highlighted HTML using the registered Highlight
// function. The language is detected from `language-<lang>` classes or the
// `data-lang` attribute.
func All() {
	doc := dom.Doc()
	codes := doc.QueryAll("pre code")
	length := codes.Length()
	for i := 0; i < length; i++ {
		el := codes.Index(i)
		cls := el.Get("className").String()
		lang := ""
		for _, c := range strings.Split(cls, " ") {
			lc := strings.ToLower(c)
			if strings.HasPrefix(lc, "language-") {
				lang = strings.TrimPrefix(lc, "language-")
				break
			}
		}
		if lang == "" {
			lang = strings.ToLower(el.Get("dataset").Get("lang").String())
		}
		code := el.Get("textContent").String()
		if res, ok := Highlight(code, lang); ok {
			el.SetHTML(res)
		}
	}
}

// HighlightAll highlights every registered code block.
func HighlightAll() { All() }

// Install exposes highlighting helpers and registers their styles.
func (p *Plugin) Install(_ *core.App) {
	js.ExposeFunc("rfwHighlight", func(_ js.Value, args []js.Value) any {
		if len(args) < 2 {
			return ""
		}
		code := args[0].String()
		lang := args[1].String()
		if res, ok := Highlight(code, lang); ok {
			return res
		}
		return ""
	})

	js.ExposeFunc("rfwHighlightAll", func(_ js.Value, _ []js.Value) any {
		All()
		return nil
	})

	doc := dom.Doc()
	style := doc.CreateElement("style")
	style.SetHTML(`.hl-kw{color:#ff7b72}.hl-tag{color:#7ee787}.hl-attr{color:#e3b341}.hl-string{color:#a5d6ff}.hl-comment{color:#8b949e;font-style:italic}.hl-var{color:#d2a8ff}.hl-cmd{color:#ffa657}`)
	doc.Query("head").Call("appendChild", style.Value)
}
