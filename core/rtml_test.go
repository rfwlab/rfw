//go:build js && wasm

package core

import (
	"strings"
	"testing"

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
