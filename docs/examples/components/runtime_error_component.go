//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/composition"
	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/runtime_error_component.rtml
var runtimeErrorComponentTpl []byte

// NewRuntimeErrorComponent demonstrates the runtime error overlay by
// triggering a JavaScript error followed by a Go panic.
func NewRuntimeErrorComponent() *core.HTMLComponent {
	cmp := composition.Wrap(core.NewComponent("RuntimeErrorComponent", runtimeErrorComponentTpl, nil))

	cmp.SetOnMount(func(*core.HTMLComponent) {
		// Schedule a JavaScript error asynchronously so the overlay captures it
		// without crashing the Go runtime.
		js.Window().Call("setTimeout", "throw new Error('js example error')", 0)

		var panicFn js.Func
		panicFn = js.FuncOf(func(this js.Value, args []js.Value) any {
			panicFn.Release()
			panic("example panic")
		})
		js.Window().Call("setTimeout", panicFn, 100)
	})

	return cmp.Unwrap()
}
