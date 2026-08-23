//go:build js && wasm

package wasmloader

import (
	"testing"

	"github.com/rfwlab/rfw/v2/js"
)

func TestLoadDelegatesToBrowserLoader(t *testing.T) {
	original := js.Get("WasmLoader")
	t.Cleanup(func() { js.Set("WasmLoader", original) })

	var gotURL string
	var gotOptions js.Value
	load := js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
		gotURL = args[0].String()
		gotOptions = args[1]
		return nil
	})
	t.Cleanup(load.Release)
	loader := js.NewDict()
	loader.Set("load", load)
	js.Set("WasmLoader", loader.Value)

	goRuntime := js.NewDict()
	Load("  /secondary.wasm?v=1  ", Options{
		Go:         goRuntime.Value,
		Color:      "#123456",
		Height:     "3px",
		Blur:       "4px",
		SkipLoader: true,
	})

	if gotURL != "/secondary.wasm?v=1" {
		t.Fatalf("url = %q, want trimmed bundle url", gotURL)
	}
	if !gotOptions.Get("go").Equal(goRuntime.Value) {
		t.Fatal("Go runtime was not forwarded")
	}
	for name, want := range map[string]string{
		"color":  "#123456",
		"height": "3px",
		"blur":   "4px",
	} {
		if got := gotOptions.Get(name).String(); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if !gotOptions.Get("skipLoader").Bool() {
		t.Fatal("skipLoader was not forwarded")
	}
}

func TestLoadIgnoresEmptyURL(t *testing.T) {
	original := js.Get("WasmLoader")
	t.Cleanup(func() { js.Set("WasmLoader", original) })

	called := false
	load := js.SafeFuncOf(func(_ js.Value, _ []js.Value) any {
		called = true
		return nil
	})
	t.Cleanup(load.Release)
	loader := js.NewDict()
	loader.Set("load", load)
	js.Set("WasmLoader", loader.Value)

	Load("   ", Options{})
	if called {
		t.Fatal("browser loader called for an empty URL")
	}
}
