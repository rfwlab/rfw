//go:build js && wasm

package core

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/state"
)

// Golden tests pinning the exact output of the production regex renderer for
// every RTML directive. Any change to substitution or escaping behavior must
// show up here as an explicit golden update.

func renderGolden(t *testing.T, name, tpl string, props map[string]any) (string, *HTMLComponent) {
	t.Helper()
	c := NewHTMLComponent(name, []byte(tpl), props)
	c.Init(nil)
	return c.Render(), c
}

func expectGolden(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("golden mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestGoldenStoreDirectives(t *testing.T) {
	st := state.NewStore("g1", state.WithModule("app"))
	st.Set("v", "<i>x</i>")
	st.Set("m", "<i>y</i>")
	tpl := `<root><p>@store:app.g1.v</p><p>@rawstore:app.g1.m</p><input value="@store:app.g1.v:w"/></root>`
	got, c := renderGolden(t, "GoldenStore", tpl, nil)
	want := fmt.Sprintf(`<root data-component-id="%s"><p><span data-store="app.g1.v">&lt;i&gt;x&lt;/i&gt;</span></p><p><span data-store-raw="app.g1.m"><i>y</i></span></p><input value="@store:app.g1.v:w"/></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenSignalDirectives(t *testing.T) {
	sig := state.NewSignal("<b>s</b>")
	tpl := `<root><p>@signal:v</p><input value="@signal:v:w"/></root>`
	got, c := renderGolden(t, "GoldenSignal", tpl, map[string]any{"v": sig})
	want := fmt.Sprintf(`<root data-component-id="%s"><p><span data-signal="v">&lt;b&gt;s&lt;/b&gt;</span></p><input value="@signal:v:w"/></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenExprDirective(t *testing.T) {
	tpl := `<root><p>@expr:n + 1</p></root>`
	got, c := renderGolden(t, "GoldenExpr", tpl, map[string]any{"n": 2})
	want := fmt.Sprintf(`<root data-component-id="%s"><p><span data-expr="expr-0">3</span></p></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenClassExprDirective(t *testing.T) {
	tpl := `<root><p class="@expr:ok ? 'on' : 'off'">t</p></root>`
	got, c := renderGolden(t, "GoldenClassExpr", tpl, map[string]any{"ok": true})
	want := fmt.Sprintf(`<root data-component-id="%s"><p class="on" data-expr-class="class-expr-0">t</p></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenPropDirectives(t *testing.T) {
	tpl := `<root><p>{{p}}</p><p>@prop:p</p><p>@rawprop:m</p><p>@prop:missing</p></root>`
	got, c := renderGolden(t, "GoldenProp", tpl, map[string]any{"p": "<u>p</u>", "m": "<u>m</u>"})
	want := fmt.Sprintf(`<root data-component-id="%s"><p>&lt;u&gt;p&lt;/u&gt;</p><p>&lt;u&gt;p&lt;/u&gt;</p><p><u>m</u></p><p>@prop:missing</p></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenIncludeDirective(t *testing.T) {
	child := NewHTMLComponent("GoldenIncChild", []byte(`<root><em>child</em></root>`), nil)
	tpl := `<root>@include:child</root>`
	c := NewHTMLComponent("GoldenInc", []byte(tpl), nil)
	c.Init(nil)
	c.AddDependency("child", child)
	got := c.Render()
	want := fmt.Sprintf(`<root data-component-id="%s"><root data-component-id="%s"><em>child</em></root>
</root>
`, c.ID, child.ID)
	expectGolden(t, got, want)
}

func TestGoldenSlotDirective(t *testing.T) {
	tpl := `<root><div>@slot:header fallback@endslot</div></root>`
	c := NewHTMLComponent("GoldenSlot", []byte(tpl), nil)
	c.Init(nil)
	c.SetSlots(map[string]any{"header": "provided"})
	got := c.Render()
	want := fmt.Sprintf(`<root data-component-id="%s"><div>provided</div></root>
`, c.ID)
	expectGolden(t, got, want)

	// Without provided content the named-slot regex drops the inline fallback
	// (the second capture group is the optional dotted modifier, not the
	// body); this golden pins that long-standing behavior.
	c2 := NewHTMLComponent("GoldenSlotFallback", []byte(tpl), nil)
	c2.Init(nil)
	got2 := c2.Render()
	want2 := fmt.Sprintf(`<root data-component-id="%s"><div></div></root>
`, c2.ID)
	expectGolden(t, got2, want2)
}

func TestGoldenForRangeDirective(t *testing.T) {
	tpl := `<root><ul>@for:i in 1..3 <li>@prop:i</li>@endfor</ul></root>`
	got, c := renderGolden(t, "GoldenForRange", tpl, nil)
	want := fmt.Sprintf(`<root data-component-id="%s"><ul> <li data-key="1">1</li> <li data-key="2">2</li> <li data-key="3">3</li></ul></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenForSliceDirective(t *testing.T) {
	st := state.NewStore("g2", state.WithModule("app"))
	st.Set("items", []any{map[string]any{"t": "<b>a</b>"}})
	tpl := `<root><ul>@for:i in store:app.g2.items <li>@prop:i.t</li>@endfor</ul></root>`
	got, c := renderGolden(t, "GoldenForSlice", tpl, nil)
	want := fmt.Sprintf(`<root data-component-id="%s"><ul> <li data-key="0">&lt;b&gt;a&lt;/b&gt;</li></ul></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenForMapDirective(t *testing.T) {
	st := state.NewStore("g3", state.WithModule("app"))
	st.Set("items", map[string]any{"k": "<b>v</b>"})
	tpl := `<root><ul>@for:k,v in store:app.g3.items <li>@prop:k=@prop:v</li>@endfor</ul></root>`
	got, c := renderGolden(t, "GoldenForMap", tpl, nil)
	want := fmt.Sprintf(`<root data-component-id="%s"><ul> <li data-key="k">k=&lt;b&gt;v&lt;/b&gt;</li></ul></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenConditionalDirective(t *testing.T) {
	tpl := "<root>\n@if:prop:v==\"1\"\nOne\n@else-if:prop:v==\"2\"\nTwo\n@else\nOther\n@endif\n</root>"
	got, c := renderGolden(t, "GoldenIf", tpl, map[string]any{"v": "2"})
	conds := []string{`@if:prop:v=="1"`, `@else-if:prop:v=="2"`, ""}
	condID := fmt.Sprintf("cond-%x", sha1.Sum([]byte(strings.Join(conds, "|"))))
	want := fmt.Sprintf("<root data-component-id=\"%s\">\n<div data-condition=\"%s\">Two\n</div></root>\n", c.ID, condID)
	expectGolden(t, got, want)
}

func TestGoldenEventDirectives(t *testing.T) {
	tpl := `<root><button @on:click:save>s</button><button @click.stop:undo>u</button></root>`
	got, c := renderGolden(t, "GoldenEvents", tpl, nil)
	want := fmt.Sprintf(`<root data-component-id="%s"><button data-on-click="save">s</button><button data-on-click="undo" data-on-click-modifiers="stop">u</button></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenRtIsDirective(t *testing.T) {
	if err := RegisterComponent("GoldenRtIsChild", func() Component {
		return NewHTMLComponent("GoldenRtIsChild", []byte(`<root><em>dyn</em></root>`), nil)
	}); err != nil && !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("register: %v", err)
	}
	tpl := `<root><div rt-is="GoldenRtIsChild"></div></root>`
	got, c := renderGolden(t, "GoldenRtIs", tpl, nil)
	child := c.Dependencies["rtis-GoldenRtIsChild-0"].(*HTMLComponent)
	want := fmt.Sprintf(`<root data-component-id="%s"><root data-component-id="%s"><em>dyn</em></root>
</div></root>
`, c.ID, child.ID)
	expectGolden(t, got, want)
}

func TestGoldenConstructorDirectives(t *testing.T) {
	tpl := `<root><div [header] class="c"></div><li [key {i.ID}]></li><span [plugin:p.badge]></span></root>`
	got, c := renderGolden(t, "GoldenConstructors", tpl, nil)
	want := fmt.Sprintf(`<root data-component-id="%s"><div data-ref="header" class="c"></div><li data-key="{i.ID}"></li><span data-plugin="p.badge"></span></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenPluginDirectives(t *testing.T) {
	RegisterPluginVar("gplug", "team", "lions")
	tpl := `<root><div @plugin:gplug.init>{plugin:gplug.team}</div></root>`
	got, c := renderGolden(t, "GoldenPlugin", tpl, nil)
	want := fmt.Sprintf(`<root data-component-id="%s"><div data-plugin-cmd="gplug.init">lions</div></root>
`, c.ID)
	expectGolden(t, got, want)
}

func TestGoldenHostDirectives(t *testing.T) {
	tpl := `<root><p>{h:count}</p><button @h:reset>r</button></root>`
	c := NewHTMLComponent("GoldenHost", []byte(tpl), map[string]any{"count": "5"})
	c.Init(nil)
	c.AddHostComponent("GoldenHostComp")
	got := c.Render()
	hash := sha1.Sum([]byte("5"))
	want := fmt.Sprintf(`<root data-component-id="%s"><p><span data-host-var="count" data-host-expected="sha1:%x">5</span></p><button data-host-cmd="reset">r</button></root>
`, c.ID, hash)
	expectGolden(t, got, want)
}
