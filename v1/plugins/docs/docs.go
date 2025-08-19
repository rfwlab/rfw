//go:build js && wasm

package docs

import (
	"encoding/json"
	"github.com/rfwlab/rfw/v1/core"
	jst "syscall/js"

	js "github.com/rfwlab/rfw/v1/js"
)

type Plugin struct {
	Sidebar string
	loader  jst.Func
}

func New(sidebar string) *Plugin {
	return &Plugin{Sidebar: sidebar}
}

func (p *Plugin) Install(a *core.App) {
	doc := js.Document()

	js.Fetch(p.Sidebar).Call("then", jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		res := args[0]
		res.Call("text").Call("then", jst.FuncOf(func(this jst.Value, args []jst.Value) any {
			js.Set("__rfwDocsSidebar", args[0].String())
			doc.Call("dispatchEvent", js.CustomEvent().New("rfwSidebar"))
			return nil
		}))
		return nil
	}))

	p.loader = jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		if len(args) < 1 {
			return nil
		}
		path := args[0].String()
		js.Fetch(path).Call("then", jst.FuncOf(func(this jst.Value, args []jst.Value) any {
			res := args[0]
			res.Call("text").Call("then", jst.FuncOf(func(this jst.Value, args []jst.Value) any {
				detail := jst.ValueOf(map[string]any{"detail": map[string]any{"path": path, "content": args[0].String()}})
				doc.Call("dispatchEvent", js.CustomEvent().New("rfwDoc", detail))
				return nil
			}))
			return nil
		}))
		return nil
	})
	js.Set("rfwLoadDoc", p.loader)
}

func (p *Plugin) Build(json.RawMessage) error { return nil }
