package dom

import (
	"regexp"
	"testing"
)

// expansion keeps the existing word grammar, including the short @on:handler
// spelling and ASCII-only names. Optimization must not accept new attributes.
func FuzzEventDirectiveMatchesExistingGrammar(f *testing.F) {
	for _, name := range []string{"", "@on:save", "@on:click:save", "@on.foo:12", "@click.enter.prevent:save_2", "@click..enter:save", "@on::save", "@ON:click:save", "@click:go-more", "@click:go\"", "@click:go<", "@click:go&", "@click:go\n", "@on:1:2", "@click:é", "@click:save:extra"} {
		f.Add(name)
	}
	grammar := regexp.MustCompile(`^@(on:)?(\w+(?:\.\w+)*):(\w+)$`)
	f.Fuzz(func(t *testing.T, name string) {
		event, handler := eventDirective(name)
		want := grammar.FindStringSubmatch(name)
		if want == nil {
			if event != "" || handler != "" {
				t.Fatalf("invalid directive %q accepted as%q/%q", name, event, handler)
			}
			return
		}
		if event != want[2] || handler != want[3] {
			t.Fatalf("directive %q became%q/%q instead of%q/%q", name, event, handler, want[2], want[3])
		}
	})
}

func TestExpandEventsKeepsShortOnAndDottedNames(t *testing.T) {
	input := `<button @on:save @on:click:save_2 @click.enter.prevent:go0 @on.foo:12>`
	want := `<button data-on-on="save" data-on-click="save_2" data-on-click="go0" data-on-click-modifiers="enter,prevent" data-on-on="12" data-on-on-modifiers="foo">`
	if got := ExpandEvents(input); got != want {
		t.Fatalf("expanded names: %q", got)
	}
}

func TestExpandEventsDoesNotCarryAttributesBetweenTags(t *testing.T) {
	input := `<button @click:a @input:b @on:keydown.enter:c>text</button><span title="@click:literal">same</span><input @change:d/>`
	want := `<button data-on-click="a" data-on-input="b" data-on-keydown="c" data-on-keydown-modifiers="enter">text</button><span title="@click:literal">same</span><input data-on-change="d"/>`
	if got := ExpandEvents(input); got != want {
		t.Fatalf("attributes crossed tag boundary: %q", got)
	}
}
