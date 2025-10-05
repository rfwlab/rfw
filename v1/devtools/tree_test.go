package devtools

import (
	"math"
	"testing"
	"time"

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
	stats         core.ComponentStats
}

func (m *mockComponent) Render() string  { return "" }
func (m *mockComponent) GetName() string { return m.name }
func (m *mockComponent) GetID() string   { return m.id }
func (m *mockComponent) Stats() core.ComponentStats {
	return m.stats
}

func TestCaptureTree(t *testing.T) {
	resetLifecycles()
	child := &mockComponent{name: "Child", id: "child"}
	root := &mockComponent{name: "Root", id: "root", Dependencies: map[string]core.Component{"child": child}}

	captureTree(root)

	if len(roots) != 1 || len(roots[0].Children) != 1 {
		t.Fatalf("tree not built correctly: %+v", roots)
	}
}

func TestTreeJSON(t *testing.T) {
	resetLifecycles()
	root := &mockComponent{name: "A", id: "a"}
	captureTree(root)
	js := treeJSON()
	if js == "" || js[0] != '[' {
		t.Fatalf("unexpected json: %s", js)
	}
}

func TestCaptureTreeNested(t *testing.T) {
	resetLifecycles()
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
	resetLifecycles()
	resetStoreUsage()
	t.Cleanup(resetStoreUsage)
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

	recordStoreBinding("root", "app", "main", "count")
	recordStoreBinding("root", "app", "main", "title")
	recordStoreBinding("child", "app", "child", "ready")

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
	if len(gotRoot.StoreBindings) != 1 {
		t.Fatalf("expected single store binding for root, got %+v", gotRoot.StoreBindings)
	}
	rootBinding := gotRoot.StoreBindings[0]
	if rootBinding.Module != "app" || rootBinding.Name != "main" {
		t.Fatalf("unexpected root binding metadata: %+v", rootBinding)
	}
	if len(rootBinding.Keys) != 2 || rootBinding.Keys[0] != "count" || rootBinding.Keys[1] != "title" {
		t.Fatalf("unexpected root binding keys: %+v", rootBinding.Keys)
	}
	if len(childNode.StoreBindings) != 1 {
		t.Fatalf("expected child binding, got %+v", childNode.StoreBindings)
	}
	if childNode.StoreBindings[0].Keys[0] != "ready" {
		t.Fatalf("unexpected child binding keys: %+v", childNode.StoreBindings[0].Keys)
	}
}

func TestStoreBindingSnapshot(t *testing.T) {
	resetStoreUsage()
	recordStoreBinding("cmp", "app", "main", "count")
	recordStoreBinding("cmp", "app", "main", "count")
	recordStoreBinding("cmp", "app", "main", "title")
	recordStoreBinding("cmp", "admin", "users", "list")
	got := snapshotStoreBindings("cmp")
	if len(got) != 2 {
		t.Fatalf("expected two store bindings, got %+v", got)
	}
	if got[0].Module != "admin" || got[0].Name != "users" {
		t.Fatalf("unexpected ordering: %+v", got)
	}
	if len(got[0].Keys) != 1 || got[0].Keys[0] != "list" {
		t.Fatalf("unexpected admin keys: %+v", got[0].Keys)
	}
	if len(got[1].Keys) != 2 || got[1].Keys[0] != "count" || got[1].Keys[1] != "title" {
		t.Fatalf("unexpected app keys: %+v", got[1].Keys)
	}
	dropStoreBindings("cmp")
	if bindings := snapshotStoreBindings("cmp"); bindings != nil {
		t.Fatalf("expected bindings cleared, got %+v", bindings)
	}
}

func TestCaptureTreeStatsAndTimeline(t *testing.T) {
	resetLifecycles()
	now := time.Now()
	appendLifecycle("root", "mount", now)
	appendLifecycle("root", "unmount", now.Add(15*time.Millisecond))
	root := &mockComponent{
		name: "Root",
		id:   "root",
		stats: core.ComponentStats{
			RenderCount:   3,
			TotalRender:   30 * time.Millisecond,
			LastRender:    12 * time.Millisecond,
			AverageRender: 10 * time.Millisecond,
			Timeline: []core.ComponentTimelineEntry{
				{Kind: "render", Timestamp: now.Add(5 * time.Millisecond), Duration: 8 * time.Millisecond},
			},
		},
	}

	captureTree(root)

	if len(roots) != 1 {
		t.Fatalf("expected single root, got %+v", roots)
	}
	got := roots[0]
	if got.Updates != 3 {
		t.Fatalf("expected render count propagated, got %d", got.Updates)
	}
	if math.Abs(got.Average-10) > 0.01 {
		t.Fatalf("expected average 10ms, got %.2f", got.Average)
	}
	if math.Abs(got.Time-12) > 0.01 {
		t.Fatalf("expected last render 12ms, got %.2f", got.Time)
	}
	if math.Abs(got.Total-30) > 0.01 {
		t.Fatalf("expected total 30ms, got %.2f", got.Total)
	}
	if len(got.Timeline) != 3 {
		t.Fatalf("expected merged timeline, got %+v", got.Timeline)
	}
	if got.Timeline[0].Kind != "mount" {
		t.Fatalf("expected mount first, got %+v", got.Timeline)
	}
	if got.Timeline[len(got.Timeline)-1].Kind != "unmount" {
		t.Fatalf("expected unmount last, got %+v", got.Timeline)
	}
}
