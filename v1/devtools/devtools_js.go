//go:build devtools && js && wasm

package devtools

import (
	"encoding/json"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/http"
	"github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/state"
	"time"
)

type plugin struct{}

var (
	root     core.Component
	treeFn   js.Func
	storeFn  js.Func
	signalFn js.Func
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
	a.RegisterStore(func(_, _, _ string, _ any) {
		if fn := js.Global().Get("RFW_DEVTOOLS_REFRESH_STORES"); fn.Type() == js.TypeFunction {
			fn.Invoke()
		}
	})
	state.SignalHook = func(int, any) {
		if fn := js.Global().Get("RFW_DEVTOOLS_REFRESH_SIGNALS"); fn.Type() == js.TypeFunction {
			fn.Invoke()
		}
	}
	http.RegisterHTTPHook(func(start bool, url string, status int, d time.Duration) {
		if obj := js.Global().Get("RFW_DEVTOOLS"); obj.Type() == js.TypeObject {
			if fn := obj.Get("network"); fn.Type() == js.TypeFunction {
				fn.Invoke(start, url, status, d.Milliseconds())
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
	storeFn = js.FuncOf(func(this js.Value, args []js.Value) any {
		b, _ := json.Marshal(state.GlobalStoreManager.Snapshot())
		return string(b)
	})
	signalFn = js.FuncOf(func(this js.Value, args []js.Value) any {
		b, _ := json.Marshal(state.SnapshotSignals())
		return string(b)
	})
	js.Global().Set("RFW_DEVTOOLS_TREE", treeFn)
	js.Global().Set("RFW_DEVTOOLS_STORES", storeFn)
	js.Global().Set("RFW_DEVTOOLS_SIGNALS", signalFn)
}
