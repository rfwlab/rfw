//go:build js && wasm

package core

import (
	"strings"
	"testing"
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
	if out != `<div class="box" data-ref="header"></div>` {
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
