//go:build js && wasm

package core

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v1/state"
)

func TestForRendersComponentList(t *testing.T) {
	state.NewStore("default", state.WithModule("app"))

	childTpl1 := []byte("<root><p>first</p></root>")
	childTpl2 := []byte("<root><p>second</p></root>")
	child1 := NewComponent("Child1", childTpl1, nil)
	child2 := NewComponent("Child2", childTpl2, nil)

	parentTpl := []byte("<root>@for:item in items @prop:item @endfor</root>")
	parent := NewComponent("Parent", parentTpl, map[string]any{"items": []Component{child1, child2}})

	html := parent.Render()
	if !strings.Contains(html, "first") || !strings.Contains(html, "second") {
		t.Fatalf("expected child components rendered: %s", html)
	}
}

func TestForRendersMapFields(t *testing.T) {
	state.NewStore("default", state.WithModule("app"))

	items := []any{
		map[string]any{"name": "Mario", "age": 30},
		map[string]any{"name": "Luigi", "age": 25},
	}

	parentTpl := []byte("<root>@for:item in items <p><b>Name:</b> @prop:item.name <b>Age:</b> @prop:item.age</p> @endfor</root>")
	parent := NewComponent("Parent", parentTpl, map[string]any{"items": items})

	html := parent.Render()
	if !strings.Contains(html, "Mario") || !strings.Contains(html, "Luigi") {
		t.Fatalf("expected names rendered: %s", html)
	}
	if strings.Contains(html, "@prop:item.name") || strings.Contains(html, "@prop:item.age") {
		t.Fatalf("placeholders not replaced: %s", html)
	}
}
