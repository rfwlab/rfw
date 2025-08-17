//go:build js && wasm

package core

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v1/state"
)

func TestNamedSlotExtraction(t *testing.T) {
	childTpl := []byte("<root>@slot:avatar<div>default</div>@endslot</root>")
	parentTpl := []byte("<root>@slot:child.avatar<img src=\"pic.png\"/>@endslot@include:child</root>")

	store := state.NewStore("test")
	parent := NewHTMLComponent("Parent", parentTpl, nil)
	parent.Init(store)
	child := NewHTMLComponent("Child", childTpl, nil)
	// child Init will be called via AddDependency
	parent.AddDependency("child", child)

	html := parent.Render()
	if strings.Contains(html, ".avatar") {
		t.Fatalf("slot placeholder not removed: %s", html)
	}
	if !strings.Contains(html, "pic.png") {
		t.Fatalf("slot content not injected: %s", html)
	}
}

func TestIncludePlaceholderPrefixCollision(t *testing.T) {
	childTpl := []byte("<root>@slot:avatar<div>fallback-avatar</div>@endslot<div>@slot<p>fallback-details</p>@endslot</div></root>")
	parentTpl := []byte("<root>@slot:card.avatar<img/>@endslot@slot:card<p>details</p>@endslot@include:card@include:cardFallback</root>")

	store := state.NewStore("test2")
	parent := NewHTMLComponent("Parent2", parentTpl, nil)
	parent.Init(store)
	card := NewHTMLComponent("Child", childTpl, nil)
	fallback := NewHTMLComponent("Child", childTpl, nil)
	parent.AddDependency("card", card)
	parent.AddDependency("cardFallback", fallback)

	html := parent.Render()
	if strings.Count(html, "<img/>") != 1 {
		t.Fatalf("expected one image only: %s", html)
	}
	if !strings.Contains(html, "fallback-avatar") || !strings.Contains(html, "fallback-details") {
		t.Fatalf("fallback content missing: %s", html)
	}
}
