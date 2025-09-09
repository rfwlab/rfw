package devtools

import (
	"testing"

	"github.com/rfwlab/rfw/v1/core"
)

type mockComponent struct {
	name         string
	id           string
	Dependencies map[string]core.Component
}

func (m *mockComponent) Render() string  { return "" }
func (m *mockComponent) GetName() string { return m.name }
func (m *mockComponent) GetID() string   { return m.id }

func TestCaptureTree(t *testing.T) {
	child := &mockComponent{name: "Child", id: "child"}
	root := &mockComponent{name: "Root", id: "root", Dependencies: map[string]core.Component{"child": child}}

	captureTree(root)

	if len(roots) != 1 || len(roots[0].Children) != 1 {
		t.Fatalf("tree not built correctly: %+v", roots)
	}
}

func TestTreeJSON(t *testing.T) {
	root := &mockComponent{name: "A", id: "a"}
	captureTree(root)
	js := treeJSON()
	if js == "" || js[0] != '[' {
		t.Fatalf("unexpected json: %s", js)
	}
}

func TestCaptureTreeNested(t *testing.T) {
	grand := &mockComponent{name: "Grand", id: "grand"}
	child := &mockComponent{name: "Child", id: "child", Dependencies: map[string]core.Component{"grand": grand}}
	root := &mockComponent{name: "Root", id: "root", Dependencies: map[string]core.Component{"child": child}}

	captureTree(root)

	if len(roots) != 1 || len(roots[0].Children) != 1 || len(roots[0].Children[0].Children) != 1 {
		t.Fatalf("tree not built correctly: %+v", roots)
	}
}
