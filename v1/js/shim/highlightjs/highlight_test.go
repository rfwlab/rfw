//go:build js && wasm

package highlightjs

import (
	"testing"

	js "github.com/rfwlab/rfw/v1/js"
)

func TestRegisterLanguage(t *testing.T) {
	dummy := js.NewDict()
	js.Set("hljs", dummy.Value)
	var name string
	dummy.Set("registerLanguage", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 2 {
			t.Fatalf("expected 2 args, got %d", len(args))
		}
		name = args[0].String()
		args[1].Invoke(js.Value{})
		return nil
	}))

	called := false
	RegisterLanguage("rtml", func(v js.Value) js.Value {
		called = true
		return js.NewDict().Value
	})

	if name != "rtml" {
		t.Fatalf("language name not passed: %s", name)
	}
	if !called {
		t.Fatal("definition callback not invoked")
	}
}
