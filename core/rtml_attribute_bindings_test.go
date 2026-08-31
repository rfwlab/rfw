//go:build js && wasm

package core

import (
	"strings"
	"testing"
)

// The validation pass is what keeps a read binding from being substituted
// inside a start tag, so it has to tell a start tag from text, a quoted value
// from an unquoted one, and the writable :w form from the read-only one.
func TestValidateAttributeBindings(t *testing.T) {
	tests := []struct {
		name     string
		template string
		binding  string
		attr     string
	}{
		{
			name:     "store in text",
			template: `<root><p>@store:app.s.k</p></root>`,
		},
		{
			name:     "rawstore and signal in text",
			template: `<root><p>@rawstore:app.s.k</p><span hidden>@signal:v</span></root>`,
		},
		{
			name:     "writable store in a form attribute",
			template: `<root><input value="@store:app.s.k:w" checked="@store:app.s.flag:w"></root>`,
		},
		{
			name:     "writable signal in a form attribute",
			template: `<root><input value="@signal:v:w"></root>`,
		},
		{
			name:     "supported attribute directives",
			template: `<root><div class="@expr:active ? 'on' : 'off'" title="@prop:label" id="{{name}}" @on:click:go></div></root>`,
		},
		{
			name:     "writable store with spaces around the equals sign",
			template: `<root><input value = "@store:app.s.k:w"></root>`,
		},
		{
			name:     "quoted angle bracket in an attribute value",
			template: `<root><a title = "a > b">@store:app.s.k</a></root>`,
		},
		{
			name:     "binding in a comment",
			template: `<root><!-- <div data-show="@store:app.s.k"> --><p>@store:app.s.k</p></root>`,
		},
		{
			name:     "binding in script text",
			template: `<root><script>if (a<b) { s = "@store:app.s.k"; }</script></root>`,
		},
		{
			name:     "read store in a quoted attribute",
			template: `<root><a data-show="@store:app.s.k" href="/x">x</a></root>`,
			binding:  "@store:app.s.k",
			attr:     "data-show",
		},
		{
			name:     "read store in an unquoted attribute",
			template: `<root><section data-show=@store:app.s.k id="s">rows</section></root>`,
			binding:  "@store:app.s.k",
			attr:     "data-show",
		},
		{
			name:     "read store in a quoted attribute with spaces around the equals sign",
			template: `<root><a data-show = "@store:app.s.k" href="/x">x</a></root>`,
			binding:  "@store:app.s.k",
			attr:     "data-show",
		},
		{
			name:     "read store in an unquoted attribute with spaces around the equals sign",
			template: "<root><section data-show\t=\n@store:app.s.k id=\"s\">rows</section></root>",
			binding:  "@store:app.s.k",
			attr:     "data-show",
		},
		{
			// a '>' inside a quoted value must not end the start tag early,
			// hiding the attributes that follow it from the scan
			name:     "read store after a quoted angle bracket",
			template: `<root><a title = "a > b" data-show="@store:app.s.k">x</a></root>`,
			binding:  "@store:app.s.k",
			attr:     "data-show",
		},
		{
			name:     "read rawstore in an attribute",
			template: `<root><a title="@rawstore:app.s.k">x</a></root>`,
			binding:  "@rawstore:app.s.k",
			attr:     "title",
		},
		{
			name:     "read signal in an attribute",
			template: `<root><a title="@signal:v">x</a></root>`,
			binding:  "@signal:v",
			attr:     "title",
		},
		{
			name:     "read store next to a writable one",
			template: `<root><input value="@store:app.s.k:w" data-show="@store:app.s.k"></root>`,
			binding:  "@store:app.s.k",
			attr:     "data-show",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAttributeBindings(tt.template)
			if tt.binding == "" {
				if err != nil {
					t.Fatalf("supported template rejected: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("unsupported attribute binding accepted")
			}
			for _, want := range []string{tt.binding, tt.attr, "@if"} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q does not mention %q", err.Error(), want)
				}
			}
		})
	}
}
