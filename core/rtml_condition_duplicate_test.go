//go:build js && wasm

package core

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/state"
)

// Two @if blocks carrying the same condition are two independent blocks. They
// used to hash to the same id, so they shared one content entry and the patch
// wrote the second block's markup into the first one.
func TestDuplicateConditionsKeepTheirOwnContent(t *testing.T) {
	store := state.NewStore("dupcond", state.WithModule("app"))
	store.Set("chrome", "on")

	tpl := []byte(`<root>
@if:store:app.dupcond.chrome == "on"
<aside data-first>first</aside>
@endif
<main>body</main>
@if:store:app.dupcond.chrome == "on"
<header data-second>second</header>
@endif
</root>`)

	c := NewHTMLComponent("DupCond", tpl, nil)
	c.SetComponent(c)
	c.Init(nil)

	html := c.Render()
	if !strings.Contains(html, "data-first") || !strings.Contains(html, "data-second") {
		t.Fatalf("both blocks should render, got: %s", html)
	}

	ids := map[string]int{}
	for _, part := range strings.Split(html, `data-condition="`)[1:] {
		ids[part[:strings.Index(part, `"`)]]++
	}
	if len(ids) != 2 {
		t.Fatalf("expected two distinct condition ids, got %v", ids)
	}
	for id, n := range ids {
		if n != 1 {
			t.Fatalf("condition id %s used %d times", id, n)
		}
	}

	first := strings.Index(html, "data-first")
	second := strings.Index(html, "data-second")
	if first > second {
		t.Fatalf("blocks rendered out of order: %s", html)
	}
}

// Re-rendering has to hand every block the same id it had before, or a patch
// after a store change lands on the wrong node.
func TestDuplicateConditionIDsAreStableAcrossRenders(t *testing.T) {
	store := state.NewStore("dupcond2", state.WithModule("app"))
	store.Set("chrome", "on")

	tpl := []byte(`<root>
@if:store:app.dupcond2.chrome == "on"
<aside>a</aside>
@endif
@if:store:app.dupcond2.chrome == "on"
<header>b</header>
@endif
</root>`)

	c := NewHTMLComponent("DupCond2", tpl, nil)
	c.SetComponent(c)
	c.Init(nil)

	ids := func() []string {
		html := c.RenderFresh()
		var out []string
		for _, part := range strings.Split(html, `data-condition="`)[1:] {
			out = append(out, part[:strings.Index(part, `"`)])
		}
		return out
	}

	before := ids()
	after := ids()
	if len(before) != 2 || len(after) != 2 {
		t.Fatalf("expected two blocks per render, got %v and %v", before, after)
	}
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("condition id %d changed between renders: %s -> %s", i, before[i], after[i])
		}
	}
}
