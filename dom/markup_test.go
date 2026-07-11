package dom

import "testing"

func TestExpandEvents(t *testing.T) {
	cases := []struct{ in, want string }{
		{`<button @on:click:save>`, `<button data-on-click="save">`},
		{`<button @click:save>`, `<button data-on-click="save">`},
		{`<input @on:keydown.enter:submit />`, `<input data-on-keydown="submit" data-on-keydown-modifiers="enter" />`},
		{`<tr @on:click:openRow data-id="3">`, `<tr data-on-click="openRow" data-id="3">`},
		{`plain text with an email@example.com stays`, `plain text with an email@example.com stays`},
	}
	for _, c := range cases {
		if got := ExpandEvents(c.in); got != c.want {
			t.Errorf("ExpandEvents(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
