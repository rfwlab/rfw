package dom

import (
	"strings"
	"testing"
)

func TestExpandEventsPreservesLiteralSSHURL(t *testing.T) {
	in := `<span>ssh://git@127.0.0.1:42601/team/repo.git</span>`
	if got := ExpandEvents(in); got != in {
		t.Fatalf("literal URL changed: %q", got)
	}
}

func TestExpandEventsPreservesAttributeValuesAndText(t *testing.T) {
	in := `<a href="ssh://git@127.0.0.1:42601/team/repo.git" title='@click:save >' data-note=@on:input:edit>@click:save <code>@on:keydown.enter:submit </code></a>`
	if got := ExpandEvents(in); got != in {
		t.Fatalf("literal data changed: %q", got)
	}
}

func TestExpandEventsPreservesCommentsAndRawText(t *testing.T) {
	in := `<!-- <button @click:save> -->` +
		`<script>const row = '<button @click:save>'; const host = 'ssh://git@127.0.0.1:42601/';</script>` +
		`<style>p::after { content: '<b @click:save>'; }</style>` +
		`<textarea><button @click:save></textarea><title><b @click:save></title>` +
		`<iframe><b @click:save></iframe><xmp><b @click:save></xmp>` +
		`<noembed><b @click:save></noembed><noframes><b @click:save></noframes>` +
		`<noscript><b @click:save></noscript>`
	if got := ExpandEvents(in); got != in {
		t.Fatalf("non-attribute content changed: %q", got)
	}
}

func TestExpandEventsResumesAfterRawTextAndKeepsAdjacentDirectives(t *testing.T) {
	in := `<SCRIPT>"</scripture><b @click:fake>"</ScRiPt><button @click:save @on:keydown.enter.prevent:submit/>`
	want := `<SCRIPT>"</scripture><b @click:fake>"</ScRiPt><button data-on-click="save" data-on-keydown="submit" data-on-keydown-modifiers="enter,prevent"/>`
	if got := ExpandEvents(in); got != want {
		t.Fatalf("expanded markup: %q", got)
	}
}

func TestExpandEventsOnlyRewritesWholeBooleanAttributes(t *testing.T) {
	in := `<button title="a > b" x@click:fake @click:save="literal" @click:go-more @on:click:real >`
	want := `<button title="a > b" x@click:fake @click:save="literal" @click:go-more data-on-click="real" >`
	if got := ExpandEvents(in); got != want {
		t.Fatalf("expanded attributes: %q", got)
	}
}

func TestExpandEventsPreservesUnclosedTagsAndPlaintext(t *testing.T) {
	in := `<p title="@click:save >" @click:edit`
	if got := ExpandEvents(in); got != in {
		t.Fatalf("unclosed tag changed: %q", got)
	}
	in = `<plaintext><button @click:save></plaintext><button @click:edit>`
	if got := ExpandEvents(in); got != in {
		t.Fatalf("plaintext changed: %q", got)
	}
}

func TestExpandEventsKeepsWhitespaceAndIsIdempotent(t *testing.T) {
	in := "<input\n\t@on:keydown.enter:submit\r\n @input:edit\t/>"
	want := "<input\n\tdata-on-keydown=\"submit\" data-on-keydown-modifiers=\"enter\"\r\n data-on-input=\"edit\"\t/>"
	got := ExpandEvents(in)
	if got != want {
		t.Fatalf("whitespace changed: %q", got)
	}
	if again := ExpandEvents(got); again != got {
		t.Fatalf("second expansion: %q", again)
	}
}

func BenchmarkExpandEventsRows(b *testing.B) {
	in := strings.Repeat(`<tr @click:open data-url="ssh://git@127.0.0.1:42601/team/repo.git"><td>@click:literal </td></tr>`, 1000)
	b.ReportAllocs()
	for b.Loop() {
		_ = ExpandEvents(in)
	}
}

func TestExpandEventsPreservesDoubleEscapedScriptText(t *testing.T) {
	in := `<script><!--<script></script><b @click:fake></script><button @click:real>`
	want := `<script><!--<script></script><b @click:fake></script><button data-on-click="real">`
	if got := ExpandEvents(in); got != want {
		t.Fatalf("script state changed: %q", got)
	}
}

func BenchmarkExpandEventsDirectiveRows(b *testing.B) {
	in := strings.Repeat(`<tr @click:open data-id="123"><td>Issue title</td><td>In progress</td></tr>`, 1000)
	b.ReportAllocs()
	for b.Loop() {
		_ = ExpandEvents(in)
	}
}
