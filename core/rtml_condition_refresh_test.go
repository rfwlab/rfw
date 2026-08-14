//go:build js && wasm

package core

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

// A block that is hidden while the state it binds to moves on has to come back
// showing the current value, not the one it carried when it left the DOM.
func TestConditionalBranchRefreshesWhenItComesBack(t *testing.T) {
	st := state.NewStore("condrefresh", state.WithModule("app"))
	st.Set("chrome", "on")
	st.Set("title", "first")

	host := dom.Doc().CreateElement("div")
	dom.Doc().Body().AppendChild(host)

	tpl := []byte(`<root>
@if:store:app.condrefresh.chrome == "on"
<header data-refresh-header>@store:app.condrefresh.title</header>
@endif
</root>`)
	c := NewHTMLComponent("CondRefresh", tpl, nil)
	c.SetComponent(c)
	c.Init(nil)
	host.SetHTML(c.Render())
	c.Mount()

	if got := dom.Query("[data-refresh-header]").Text(); got != "first" {
		t.Fatalf("initial header = %q", got)
	}

	st.Set("chrome", "off")
	waitForRenderFlush()
	if el := dom.Query("[data-refresh-header]"); !el.IsNull() {
		t.Fatal("header should be gone while the condition is false")
	}

	// the title moves while the block is out of the DOM
	st.Set("title", "second")
	st.Set("chrome", "on")
	waitForRenderFlush()

	el := dom.Query("[data-refresh-header]")
	if el.IsNull() {
		t.Fatal("header did not come back")
	}
	if got := strings.TrimSpace(el.Text()); got != "second" {
		t.Fatalf("header came back stale: %q", got)
	}
}
