//go:build js && wasm

package highlight

import (
	"encoding/json"
	"strings"

	"github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	js "github.com/rfwlab/rfw/v1/js"
)

type Plugin struct{}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Build(json.RawMessage) error { return nil }

func (p *Plugin) Install(a *core.App) {
	js.ExposeFunc("rfwHighlight", func(this js.Value, args []js.Value) any {
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

	js.ExposeFunc("rfwHighlightAll", func(this js.Value, args []js.Value) any {
		codes := dom.QueryAll("pre code")
		length := codes.Length()
		for i := 0; i < length; i++ {
			el := codes.Index(i)
			cls := el.Get("className").String()
			lang := ""
			for _, c := range strings.Split(cls, " ") {
				if strings.HasPrefix(c, "language-") {
					lang = strings.TrimPrefix(c, "language-")
					break
				}
			}
			if lang == "" {
				lang = el.Get("dataset").Get("lang").String()
			}
			code := el.Get("textContent").String()
			if res, ok := Highlight(code, lang); ok {
				dom.SetInnerHTML(el, res)
			}
		}
		return nil
	})

	style := dom.CreateElement("style")
	dom.SetInnerHTML(style, `.hl-kw{color:#ff7b72}.hl-tag{color:#7ee787}.hl-attr{color:#e3b341}.hl-string{color:#a5d6ff}.hl-comment{color:#8b949e;font-style:italic}.hl-var{color:#d2a8ff}.hl-cmd{color:#ffa657}`)
	js.Doc().Get("head").Call("appendChild", style)
}
