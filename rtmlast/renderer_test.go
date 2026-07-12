//go:build js && wasm

package rtmlast

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/state"
)

// RenderNodes follows the same escape-by-default policy as the production
// renderer in core: {{var}}, @prop, @store and @signal output is text.
func TestRenderNodesEscapesValues(t *testing.T) {
	mgr := state.NewStoreManager()
	st := mgr.NewStore("s", state.WithModule("app"))
	st.Set("v", "<i>store</i>")

	nodes, err := Parse("{{name}}\n@prop:name\n@store:app.s.v\n")
	if err != nil {
		t.Fatal(err)
	}
	ctx := &RenderContext{
		Props:    map[string]any{"name": "<b>x</b>"},
		StoreMgr: mgr,
	}
	out := RenderNodes(nodes, ctx)
	if strings.Contains(out, "<b>x</b>") || strings.Contains(out, "<i>store</i>") {
		t.Fatalf("markup injected: %s", out)
	}
	if !strings.Contains(out, "&lt;b&gt;x&lt;/b&gt;") {
		t.Fatalf("var/prop not escaped: %s", out)
	}
	if !strings.Contains(out, "&lt;i&gt;store&lt;/i&gt;") {
		t.Fatalf("store not escaped: %s", out)
	}
}

func TestRenderNodesEscapesSignals(t *testing.T) {
	sig := state.NewSignal("<img src=x>")
	nodes, err := Parse("@signal:v")
	if err != nil {
		t.Fatal(err)
	}
	out := RenderNodes(nodes, &RenderContext{Props: map[string]any{"v": sig}})
	if !strings.Contains(out, "&lt;img src=x&gt;") {
		t.Fatalf("signal not escaped: %s", out)
	}
}
