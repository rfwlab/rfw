//go:build devtools && js && wasm

package devtools

import (
	"encoding/json"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/js"
)

type plugin struct{}

var (
	root   core.Component
	treeFn js.Func
)

func (plugin) Build(json.RawMessage) error { return nil }
func (plugin) Install(a *core.App) {
	a.RegisterLifecycle(func(c core.Component) {
		if root == nil {
			root = c
		}
		if root != nil {
			captureTree(root)
			if fn := js.Global().Get("RFW_DEVTOOLS_REFRESH"); fn.Type() == js.TypeFunction {
				fn.Invoke()
			}
		}
	}, func(c core.Component) {
		if root == c {
			resetTree()
			root = nil
		} else if root != nil {
			captureTree(root)
			if fn := js.Global().Get("RFW_DEVTOOLS_REFRESH"); fn.Type() == js.TypeFunction {
				fn.Invoke()
			}
		}
	})
	a.RegisterRouter(func(_ string) {
		if root != nil {
			captureTree(root)
			if fn := js.Global().Get("RFW_DEVTOOLS_REFRESH"); fn.Type() == js.TypeFunction {
				fn.Invoke()
			}
		}
	})
}

func init() {
	core.RegisterPlugin(plugin{})
	treeFn = js.FuncOf(func(this js.Value, args []js.Value) any {
		if root != nil {
			captureTree(root)
		}
		return treeJSON()
	})
	js.Global().Set("RFW_DEVTOOLS_TREE", treeFn)
}
