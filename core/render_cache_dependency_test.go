//go:build js && wasm

package core

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/state"
)

// A dependency bound to a store must not be frozen by the parent's render
// cache: the cache key knows nothing about store state, so re-rendering the
// parent has to re-render the included subtree too.
func TestDependencyRenderFollowsStore(t *testing.T) {
	store := state.NewStore("depcache", state.WithModule("app"))
	store.Set("chrome", "on")
	defer state.GlobalStoreManager.UnregisterStore("app", "depcache")

	child := NewHTMLComponent("CacheChild", []byte(`<root>
@if:store:app.depcache.chrome == "on"
<span id="child-block">visible</span>
@endif
</root>`), nil)
	child.SetComponent(child)
	child.Init(nil)

	parent := NewHTMLComponent("CacheParent", []byte(`<root><div>@include:child</div></root>`), nil)
	parent.SetComponent(parent)
	parent.AddDependency("child", child)
	parent.Init(nil)

	if html := parent.Render(); !strings.Contains(html, "child-block") {
		t.Fatalf("first render missed the true branch: %s", html)
	}

	store.Set("chrome", "off")

	if html := parent.RenderFresh(); strings.Contains(html, "child-block") {
		t.Fatalf("dependency kept its cached render after the store changed: %s", html)
	}
}
