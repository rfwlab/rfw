//go:build js && wasm

package docs

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/markdown"
	"github.com/rfwlab/rfw/v1/plugins/seo"
)

type Plugin struct {
	Sidebar    string
	loader     js.Func
	disableSEO bool
}

func New(sidebar string, disableSEO ...bool) *Plugin {
	sidebar = fmt.Sprintf("%s?%s", sidebar, time.Now().Unix())
	p := &Plugin{Sidebar: sidebar}
	if len(disableSEO) > 0 {
		p.disableSEO = disableSEO[0]
	}
	return p
}

func (p *Plugin) Name() string { return "docs" }

func (p *Plugin) Optional() []core.Plugin {
	if p.disableSEO {
		return nil
	}
	return []core.Plugin{seo.New()}
}

func (p *Plugin) Install(a *core.App) {
	doc := js.Document()

	js.Fetch(p.Sidebar).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		res := args[0]
		res.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			js.Set("__rfwDocsSidebar", args[0].String())
			doc.Call("dispatchEvent", js.CustomEvent().New("rfwSidebar"))
			return nil
		}))
		return nil
	}))

	p.loader = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return nil
		}
		path := args[0].String()
		js.Fetch(path).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			res := args[0]
			res.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				content := args[0].String()
				mhs := markdown.Headings(content)
				headings := make([]any, len(mhs))
				for i, h := range mhs {
					headings[i] = map[string]any{"text": h.Text, "depth": h.Depth, "id": h.ID}
				}
				doc.Call("dispatchEvent", js.CustomEvent().New("rfwDoc", map[string]any{"detail": map[string]any{"path": path, "content": content, "headings": headings}}))
				return nil
			}))
			return nil
		}))
		return nil
	})
	js.Set("rfwLoadDoc", p.loader)
}

func (p *Plugin) Build(json.RawMessage) error { return nil }
