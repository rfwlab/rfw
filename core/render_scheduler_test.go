//go:build js && wasm

package core

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

func TestRenderSchedulerCoalescesStoreBurst(t *testing.T) {
	store := state.NewStore("scheduler_burst", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "scheduler_burst")
	store.Set("items", []any{map[string]any{"label": "0"}})

	component := mountForComponent(t, "SchedulerBurst", []byte(`<root><ul>
@for:item in store:app.scheduler_burst.items
<li data-key="row">@prop:item.label</li>
@endfor
</ul></root>`))
	defer component.Unmount()
	before := component.Stats().RenderCount

	for i := 1; i <= 20; i++ {
		store.Set("items", []any{map[string]any{"label": fmt.Sprintf("%d", i)}})
	}
	waitForRenderFlush()

	if got := component.Stats().RenderCount - before; got != 1 {
		t.Fatalf("render burst produced %d renders, want 1", got)
	}
	if got := dom.ComponentRoot(component.ID).Query("li").Text(); got != "20" {
		t.Fatalf("scheduled render used %q, want latest value 20", got)
	}
}

func TestRenderSchedulerCancelsUnmountedComponent(t *testing.T) {
	store := state.NewStore("scheduler_unmount", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "scheduler_unmount")
	store.Set("items", []any{map[string]any{"label": "before"}})

	component := mountForComponent(t, "SchedulerUnmount", []byte(`<root><ul>
@for:item in store:app.scheduler_unmount.items
<li>@prop:item.label</li>
@endfor
</ul></root>`))
	before := component.Stats().RenderCount
	store.Set("items", []any{map[string]any{"label": "after"}})
	component.Unmount()
	waitForRenderFlush()

	if got := component.Stats().RenderCount; got != before {
		t.Fatalf("unmounted component rendered: before=%d after=%d", before, got)
	}
}

func TestRenderSchedulerContinuesAfterDOMFailure(t *testing.T) {
	store := state.NewStore("scheduler_recovery", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "scheduler_recovery")
	store.Set("items", []any{map[string]any{"label": "before"}})
	component := mountForComponent(t, "SchedulerRecovery", []byte(`<root><ul>
@for:item in store:app.scheduler_recovery.items
<li>@prop:item.label</li>
@endfor
</ul></root>`))
	defer component.Unmount()

	previousHook := dom.TemplateHook
	dom.TemplateHook = func(string, string) { panic("test hook") }
	store.Set("items", []any{map[string]any{"label": "failed-cycle"}})
	waitForRenderFlush()
	dom.TemplateHook = nil
	defer func() { dom.TemplateHook = previousHook }()

	store.Set("items", []any{map[string]any{"label": "recovered"}})
	waitForRenderFlush()
	if got := dom.ComponentRoot(component.ID).Query("li").Text(); got != "recovered" {
		t.Fatalf("scheduler wedged after recovered failure: got %q", got)
	}
}

func TestRenderSchedulerCommitsParentBeforeChild(t *testing.T) {
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	child := NewHTMLComponent("ScheduledChild", []byte(`<root><span>child</span></root>`), nil)
	child.SetComponent(child)
	child.Init(nil)
	parent := NewHTMLComponent("ScheduledParent", []byte(`<root><main>parent</main>@include:child</root>`), nil)
	parent.SetComponent(parent)
	parent.AddDependency("child", child)
	parent.Init(nil)
	dom.UpdateDOM(parent.ID, parent.Render())
	parent.Mount()
	defer parent.Unmount()

	previousHook := dom.TemplateHook
	var commits []string
	dom.TemplateHook = func(componentID, _ string) {
		if componentID == parent.ID || componentID == child.ID {
			commits = append(commits, componentID)
		}
	}
	defer func() { dom.TemplateHook = previousHook }()

	child.requestRender()
	parent.requestRender()
	waitForRenderFlush()
	if want := []string{parent.ID, child.ID}; !reflect.DeepEqual(commits, want) {
		t.Fatalf("commit order = %v, want parent then child %v", commits, want)
	}
}

func TestRenderSchedulerCommitsSiblingsByStableID(t *testing.T) {
	var commits []string
	active := func() bool { return true }
	requestScheduledRender(renderJob{
		id:     "sibling-b",
		depth:  1,
		active: active,
		render: func() { commits = append(commits, "sibling-b") },
	})
	requestScheduledRender(renderJob{
		id:     "sibling-a",
		depth:  1,
		active: active,
		render: func() { commits = append(commits, "sibling-a") },
	})
	waitForRenderFlush()

	if want := []string{"sibling-a", "sibling-b"}; !reflect.DeepEqual(commits, want) {
		t.Fatalf("sibling commit order = %v, want stable ID order %v", commits, want)
	}
}

func waitForRenderFlush() {
	time.Sleep(20 * time.Millisecond)
}
