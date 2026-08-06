//go:build js && wasm

package core

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/state"
)

// Tests complex conditional scenarios including @else-if and nested blocks.
func TestElseIfRendering(t *testing.T) {
	c := &HTMLComponent{Props: map[string]any{"val": "2"}, conditionContents: make(map[string]ConditionContent)}
	template := `
@if:prop:val=="1"
One
@else-if:prop:val=="2"
Two
@else
Other
@endif`

	out := replaceConditionals(template, c)
	if strings.Contains(out, "One") || strings.Contains(out, "Other") {
		t.Fatalf("unexpected branches rendered: %s", out)
	}
	if !strings.Contains(out, "Two") {
		t.Fatalf("expected 'Two' branch, got %s", out)
	}
}

func TestNestedConditionals(t *testing.T) {
	props := map[string]any{"outer": "yes", "inner": "maybe"}
	c := &HTMLComponent{Props: props, conditionContents: make(map[string]ConditionContent)}
	template := `
@if:prop:outer=="yes"
Start
    @if:prop:inner=="yes"
        InnerYes
    @else-if:prop:inner=="maybe"
        InnerMaybe
    @else
        InnerNo
    @endif
@else
    OuterNo
@endif`

	out := replaceConditionals(template, c)
	if !strings.Contains(out, "Start") || !strings.Contains(out, "InnerMaybe") {
		t.Fatalf("nested conditions not rendered as expected: %s", out)
	}
	if strings.Contains(out, "InnerYes") || strings.Contains(out, "InnerNo") || strings.Contains(out, "OuterNo") {
		t.Fatalf("unexpected branches present: %s", out)
	}
}

// Tests constructor decorators for refs and keyed lists.
func TestReplaceConstructors(t *testing.T) {
	tpl := `<div [header] class="box"></div>`
	out := replaceConstructors(tpl)
	if !strings.Contains(out, `data-ref="header"`) || !strings.Contains(out, `class="box"`) {
		t.Fatalf("unexpected constructor replacement: %s", out)
	}

	tpl = `<li [key {item.ID}]></li>`
	out = replaceConstructors(tpl)
	if out != `<li data-key="{item.ID}"></li>` {
		t.Fatalf("expected data-key constructor, got %s", out)
	}
}

// Tests plugin placeholders for variables, commands and constructors.
func TestPluginPlaceholders(t *testing.T) {
	RegisterPluginVar("soccer", "team", "lions")
	tpl := `<div @plugin:soccer.init>{plugin:soccer.team}</div>`
	out := replacePluginPlaceholders(tpl)
	if !strings.Contains(out, "lions") {
		t.Fatalf("plugin variable not replaced: %s", out)
	}
	if !strings.Contains(out, `data-plugin-cmd="soccer.init"`) {
		t.Fatalf("plugin command not replaced: %s", out)
	}

	tpl = `<span [plugin:soccer.badge]></span>`
	out = replaceConstructors(tpl)
	if !strings.Contains(out, `data-plugin="soccer.badge"`) {
		t.Fatalf("plugin constructor not replaced: %s", out)
	}
}

// A @for bound to a store key that was never set must render nothing: the raw
// template row used to leak into the DOM and keyed patches never removed it.
func TestForUnsetStoreKeyRendersNothing(t *testing.T) {
	state.NewStore("fornil", state.WithModule("app"))
	tpl := []byte(`<root><div>@for:i in store:app.fornil.items
<span data-key="x">@prop:i.title</span>
@endfor</div></root>`)
	c := NewHTMLComponent("ForNil", tpl, nil)
	c.Init(nil)
	html := c.Render()
	if strings.Contains(html, "@prop") || strings.Contains(html, "@for") {
		t.Fatalf("unset store key leaked template markup: %s", html)
	}
}

// Substituted values are HTML-escaped by default; @rawprop opts into markup.
func TestForFieldsEscapedByDefault(t *testing.T) {
	st := state.NewStore("escfor", state.WithModule("app"))
	st.Set("items", []any{map[string]any{"txt": "<img src=x>", "markup": "<b>ok</b>"}})
	tpl := []byte(`<root><div>@for:i in store:app.escfor.items
<span data-key="k">@prop:i.txt|@rawprop:i.markup</span>
@endfor</div></root>`)
	c := NewHTMLComponent("EscFor", tpl, nil)
	c.Init(nil)
	html := c.Render()
	if !strings.Contains(html, "&lt;img src=x&gt;") {
		t.Fatalf("field not escaped: %s", html)
	}
	if !strings.Contains(html, "<b>ok</b>") {
		t.Fatalf("rawprop escaped: %s", html)
	}
}

// @signal values are escaped by default: a signal carrying user-supplied
// markup must render as text, not execute as HTML.
func TestSignalEscapedByDefault(t *testing.T) {
	sig := state.NewSignal("<img src=x onerror=alert(1)>")
	tpl := []byte(`<root><div>@signal:v</div></root>`)
	c := NewHTMLComponent("EscSignal", tpl, map[string]any{"v": sig})
	c.Init(nil)
	html := c.Render()
	if !strings.Contains(html, "&lt;img src=x onerror=alert(1)&gt;") {
		t.Fatalf("signal not escaped: %s", html)
	}
	if strings.Contains(html, "<img") {
		t.Fatalf("signal injected markup: %s", html)
	}
}

// @expr output is escaped by default, matching @store/@prop policy.
func TestExprEscapedByDefault(t *testing.T) {
	tpl := []byte(`<root><div>@expr:msg</div></root>`)
	c := NewHTMLComponent("EscExpr", tpl, map[string]any{"msg": "<b>bad</b>"})
	c.Init(nil)
	html := c.Render()
	if !strings.Contains(html, "&lt;b&gt;bad&lt;/b&gt;") {
		t.Fatalf("expr not escaped: %s", html)
	}
	if strings.Contains(html, "<b>bad</b>") {
		t.Fatalf("expr injected markup: %s", html)
	}
}

// Signal updates go through textContent so later values cannot inject HTML.
func TestSignalUpdateUsesTextContent(t *testing.T) {
	sig := state.NewSignal("safe")
	tpl := []byte(`<root><div>@signal:v</div></root>`)
	c := NewHTMLComponent("EscSignalUpdate", tpl, map[string]any{"v": sig})
	c.Init(nil)
	html := c.Render()
	container := dom.Doc().CreateElement("div")
	container.Set("innerHTML", html)
	dom.Doc().Get("body").Call("appendChild", container.Value)
	defer container.Call("remove")
	sig.Set("<img src=x onerror=alert(1)>")
	root := dom.ComponentRoot(c.ID)
	node := root.Query(`[data-signal="v"]`)
	if node.IsNull() || node.IsUndefined() {
		t.Fatalf("signal binding not found")
	}
	inner := node.Get("innerHTML").String()
	if !strings.Contains(inner, "&lt;img") {
		t.Fatalf("signal update not applied as text: %s", inner)
	}
}

// {{prop}} substitutions are escaped by default like @prop; @rawprop stays
// the trusted markup escape hatch.
func TestCurlyPropEscapedByDefault(t *testing.T) {
	tpl := []byte(`<root><div>{{msg}}|@rawprop:markup</div></root>`)
	c := NewHTMLComponent("EscCurly", tpl, map[string]any{
		"msg":    "<script>alert(1)</script>",
		"markup": "<b>ok</b>",
	})
	c.Init(nil)
	html := c.Render()
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("curly prop not escaped: %s", html)
	}
	if !strings.Contains(html, "<b>ok</b>") {
		t.Fatalf("rawprop escaped: %s", html)
	}
}

// @for over a map collection escapes keys, scalar values and map fields just
// like the slice path; @rawprop opts into trusted markup.
func TestForMapFieldsEscapedByDefault(t *testing.T) {
	st := state.NewStore("escformap", state.WithModule("app"))
	st.Set("items", map[string]any{
		"<key>": map[string]any{"txt": "<img src=x>", "markup": "<b>ok</b>"},
	})
	tpl := []byte(`<root><div>@for:k,v in store:app.escformap.items
<span data-key="m">@prop:k|@prop:v.txt|@rawprop:v.markup</span>
@endfor</div></root>`)
	c := NewHTMLComponent("EscForMap", tpl, nil)
	c.Init(nil)
	html := c.Render()
	if !strings.Contains(html, "&lt;key&gt;") {
		t.Fatalf("map key not escaped: %s", html)
	}
	if !strings.Contains(html, "&lt;img src=x&gt;") {
		t.Fatalf("map field not escaped: %s", html)
	}
	if !strings.Contains(html, "<b>ok</b>") {
		t.Fatalf("rawprop escaped: %s", html)
	}
}

// @for over a map of scalars escapes values by default.
func TestForMapScalarsEscapedByDefault(t *testing.T) {
	st := state.NewStore("escformapscalar", state.WithModule("app"))
	st.Set("items", map[string]any{"a": "<i>x</i>"})
	tpl := []byte(`<root><div>@for:k,v in store:app.escformapscalar.items
<span data-key="s">@prop:v</span>
@endfor</div></root>`)
	c := NewHTMLComponent("EscForMapScalar", tpl, nil)
	c.Init(nil)
	html := c.Render()
	if !strings.Contains(html, "&lt;i&gt;x&lt;/i&gt;") {
		t.Fatalf("map scalar not escaped: %s", html)
	}
}

// An @if body holding sibling elements renders them inside one container, in
// order. Callers style around that wrapper, so it stays part of the contract.
func TestConditionalWrapsSiblingsInOneContainer(t *testing.T) {
	tpl := []byte(`<root><div>
@if:prop:show=="yes"
<span class="a">A</span>
<span class="b">B</span>
@endif
</div></root>`)
	c := NewHTMLComponent("CondSiblings", tpl, map[string]any{"show": "yes"})
	c.Init(nil)
	html := c.Render()
	if n := strings.Count(html, "data-condition="); n != 1 {
		t.Fatalf("expected one conditional container, got %d: %s", n, html)
	}
	open := strings.Index(html, "data-condition=")
	closeIdx := strings.Index(html[open:], "</div>")
	block := html[open : open+closeIdx]
	a, b := strings.Index(block, `class="a"`), strings.Index(block, `class="b"`)
	if a < 0 || b < 0 {
		t.Fatalf("siblings not both inside the container: %s", block)
	}
	if a > b {
		t.Fatalf("siblings reordered: %s", block)
	}
}

// An @if inside a @for body is expanded per row and every row keeps its own
// conditional block, so one row's branch cannot patch another's.
func TestForBodyConditionalPerRow(t *testing.T) {
	st := state.NewStore("forcond", state.WithModule("app"))
	st.Set("items", []any{
		map[string]any{"name": "one"},
		map[string]any{"name": "two"},
	})
	tpl := []byte(`<root><div>
@for:i in store:app.forcond.items
<article data-key="k">
@if:prop:flag=="on"
<span class="x">X</span>
<span class="y">@prop:i.name</span>
@endif
</article>
@endfor
</div></root>`)
	c := NewHTMLComponent("ForCond", tpl, map[string]any{"flag": "on"})
	c.Init(nil)
	html := c.Render()
	if n := strings.Count(html, "data-condition="); n != 2 {
		t.Fatalf("expected one conditional per row, got %d: %s", n, html)
	}
	if strings.Count(html, `class="x"`) != 2 {
		t.Fatalf("branch not rendered on every row: %s", html)
	}
	for _, name := range []string{"one", "two"} {
		if !strings.Contains(html, ">"+name+"<") {
			t.Fatalf("row value %q missing inside the conditional: %s", name, html)
		}
	}
	if strings.Contains(html, "@prop") || strings.Contains(html, "@if") {
		t.Fatalf("directives leaked into the output: %s", html)
	}
}

// A loop body carrying no directive at all still renders one row per item.
func TestForBodyWithoutDirectives(t *testing.T) {
	st := state.NewStore("forplain", state.WithModule("app"))
	st.Set("items", []any{
		map[string]any{"name": "a"},
		map[string]any{"name": "b"},
		map[string]any{"name": "c"},
	})
	tpl := []byte(`<root><ul>
@for:i in store:app.forplain.items
<li class="row"><b>@prop:i.name</b></li>
@endfor
</ul></root>`)
	c := NewHTMLComponent("ForPlain", tpl, nil)
	c.Init(nil)
	html := c.Render()
	if n := strings.Count(html, `class="row"`); n != 3 {
		t.Fatalf("expected three rows, got %d: %s", n, html)
	}
	for _, name := range []string{"a", "b", "c"} {
		if !strings.Contains(html, ">"+name+"<") {
			t.Fatalf("row %q missing: %s", name, html)
		}
	}
}

// @store values are escaped by default; @rawstore injects trusted markup.
func TestStoreEscapedByDefault(t *testing.T) {
	st := state.NewStore("escstore", state.WithModule("app"))
	st.Set("v", "<i>x</i>")
	st.Set("m", "<i>y</i>")
	tpl := []byte(`<root><div>@store:app.escstore.v @rawstore:app.escstore.m</div></root>`)
	c := NewHTMLComponent("EscStore", tpl, nil)
	c.Init(nil)
	html := c.Render()
	if !strings.Contains(html, "&lt;i&gt;x&lt;/i&gt;") {
		t.Fatalf("store not escaped: %s", html)
	}
	if !strings.Contains(html, "<i>y</i>") {
		t.Fatalf("rawstore escaped: %s", html)
	}
}

// @for has always split store paths by hand, so it accepted a hyphenated store
// name while @store, @rawstore and conditions matched \w and quietly left the
// directive in the markup. A binding that renders as its own source text is
// worse than one that fails, so the three agree on the character set now.
func TestHyphenatedStorePathsResolve(t *testing.T) {
	st := state.NewStore("hyphen-store", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "hyphen-store")
	st.Set("first-name", "Mirko")
	st.Set("markup", "<i>y</i>")

	tpl := []byte(`<root>
@if:store:app.hyphen-store.first-name == "Mirko"
<div data-hyphen>@store:app.hyphen-store.first-name @rawstore:app.hyphen-store.markup</div>
@endif
</root>`)
	c := NewHTMLComponent("HyphenStore", tpl, nil)
	c.Init(nil)
	html := c.Render()

	if strings.Contains(html, "@store:") || strings.Contains(html, "@rawstore:") {
		t.Fatalf("hyphenated store path left in the markup: %s", html)
	}
	if !strings.Contains(html, "data-hyphen") {
		t.Fatalf("condition on a hyphenated store did not select the branch: %s", html)
	}
	if !strings.Contains(html, "Mirko") {
		t.Fatalf("hyphenated store value missing: %s", html)
	}
	if !strings.Contains(html, "<i>y</i>") {
		t.Fatalf("hyphenated rawstore escaped: %s", html)
	}
}

// The two-way binding is parsed again from the rendered attribute, so the DOM
// side has to accept the same paths the template does.
func TestHyphenatedStoreWriteBindingUpdatesStore(t *testing.T) {
	st := state.NewStore("hyphen-write", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "hyphen-write")
	st.Set("query-text", "")

	tpl := []byte(`<root><input data-hyphen-input value="@store:app.hyphen-write.query-text:w"></root>`)
	c := mountForComponent(t, "HyphenWrite", tpl)
	defer c.Unmount()

	input := dom.Query("[data-hyphen-input]")
	input.Set("value", "typed")
	input.Call("dispatchEvent", js.CustomEvent().New("input"))

	if got := st.Get("query-text"); got != "typed" {
		t.Fatalf("store key = %v, want %q", got, "typed")
	}
}

// A condition's dependencies come from a regex of their own, so a hyphenated
// store also has to register the subscription that repaints the branch when
// the value moves. Without it the block renders once and then freezes.
func TestConditionOnHyphenatedStoreReacts(t *testing.T) {
	st := state.NewStore("hyphen-cond", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "hyphen-cond")
	st.Set("chrome-state", "on")

	tpl := []byte(`<root>
@if:store:app.hyphen-cond.chrome-state == "on"
<header data-hyphen-header>shown</header>
@endif
</root>`)
	c := mountForComponent(t, "HyphenCond", tpl)
	defer c.Unmount()

	if dom.Query("[data-hyphen-header]").IsNull() {
		t.Fatal("header not rendered while the condition holds")
	}

	st.Set("chrome-state", "off")
	waitForRenderFlush()
	if el := dom.Query("[data-hyphen-header]"); !el.IsNull() {
		t.Fatal("header stayed after the hyphenated store key changed")
	}
}
