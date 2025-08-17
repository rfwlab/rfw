//go:build js && wasm

package core

import (
	"strings"
	"testing"
)

// Tests complex conditional scenarios including @else-if and nested blocks.
func TestElseIfRendering(t *testing.T) {
	c := &HTMLComponent{Props: map[string]interface{}{"val": "2"}, conditionContents: make(map[string]ConditionContent)}
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
	props := map[string]interface{}{"outer": "yes", "inner": "maybe"}
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
