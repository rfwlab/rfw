//go:build js && wasm

package virtual

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
)

func TestVirtualListRendersVisibleItems(t *testing.T) {
	c := dom.CreateElement("div")
	c.Set("id", "vlist")
	c.Get("style").Set("height", "60px")
	c.Get("style").Set("overflow", "auto")
	dom.Doc().Body().AppendChild(c)
	defer c.Call("remove")

	v := NewVirtualList("vlist", 100, 20, func(i int) string {
		return "<div class='it'>row</div>"
	})
	defer v.Destroy()

	html := c.HTML()
	if !strings.Contains(html, "class='it'") && !strings.Contains(html, `class="it"`) {
		t.Fatalf("expected rendered items, got %q", html)
	}
}

func TestVirtualListMissingContainer(t *testing.T) {
	v := NewVirtualList("does-not-exist", 10, 20, func(i int) string { return "" })
	if v == nil {
		t.Fatalf("expected VirtualList instance")
	}
	v.Destroy()
}
