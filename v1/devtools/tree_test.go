package devtools

import (
	"testing"

	"github.com/rfwlab/rfw/v1/core"
)

type mockComponent struct {
	name          string
	id            string
	Dependencies  map[string]core.Component
	Props         map[string]any
	Slots         map[string]any
	Signals       map[string]any
	Store         storeInspector
	HostComponent string
	Updates       int
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

type fakeStore struct {
	module string
	name   string
	state  map[string]any
}

func (s *fakeStore) Snapshot() map[string]any {
	dup := make(map[string]any, len(s.state))
	for k, v := range s.state {
		dup[k] = v
	}
	return dup
}

func (s *fakeStore) Module() string { return s.module }
func (s *fakeStore) Name() string   { return s.name }

func TestCaptureTreeMetadata(t *testing.T) {
	child := &mockComponent{name: "Child", id: "child", Updates: 2}
	root := &mockComponent{
		name: "Root",
		id:   "root",
		Dependencies: map[string]core.Component{
			"child": child,
		},
		Props: map[string]any{"title": "hello", "count": 3},
		Slots: map[string]any{"header": "value"},
		Signals: map[string]any{
			"selected": "item",
		},
		Store: &fakeStore{
			module: "app",
			name:   "main",
			state: map[string]any{
				"count": 7,
			},
		},
		HostComponent: "Widget",
		Updates:       5,
	}

	captureTree(root)

	if len(roots) != 1 {
		t.Fatalf("expected single root, got %+v", roots)
	}
	gotRoot := roots[0]
	if gotRoot.Props["title"] != "hello" {
		t.Fatalf("expected props copied, got %+v", gotRoot.Props)
	}
	if gotRoot.Slots["header"] != "value" {
		t.Fatalf("expected slots copied, got %+v", gotRoot.Slots)
	}
	if gotRoot.Signals["selected"] != "item" {
		t.Fatalf("expected signals copied, got %+v", gotRoot.Signals)
	}
	if gotRoot.Host != "Widget" {
		t.Fatalf("expected host set, got %q", gotRoot.Host)
	}
	if gotRoot.Store == nil || gotRoot.Store.Module != "app" || gotRoot.Store.Name != "main" {
		t.Fatalf("unexpected store snapshot: %+v", gotRoot.Store)
	}
	if gotRoot.Store.State["count"] != int64(7) && gotRoot.Store.State["count"] != float64(7) {
		t.Fatalf("expected store state value, got %+v", gotRoot.Store.State["count"])
	}
	if len(gotRoot.Children) != 1 {
		t.Fatalf("expected child node, got %+v", gotRoot.Children)
	}
	childNode := gotRoot.Children[0]
	if childNode.Owner != "Root" {
		t.Fatalf("expected owner to be root, got %q", childNode.Owner)
	}
	if childNode.Updates != 2 {
		t.Fatalf("expected child updates, got %d", childNode.Updates)
	}
}
