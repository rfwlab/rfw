package host

import (
	"strings"
	"testing"
)

func TestSpan(t *testing.T) {
	got := Span("Visit", 1)
	if !strings.Contains(got, `data-host-var="Visit"`) {
		t.Fatalf("missing data-host-var: %s", got)
	}
	if !strings.Contains(got, `data-host-expected="sha1:`) {
		t.Fatalf("missing data-host-expected: %s", got)
	}
	if !strings.Contains(got, ">1</span>") {
		t.Fatalf("missing value: %s", got)
	}
	if !strings.HasPrefix(got, "<span ") {
		t.Fatalf("wrong tag: %s", got)
	}
}

func TestDiv(t *testing.T) {
	got := Div("message", "hello")
	if !strings.Contains(got, `data-host-var="message"`) {
		t.Fatalf("missing data-host-var: %s", got)
	}
	if !strings.Contains(got, ">hello</div>") {
		t.Fatalf("missing value: %s", got)
	}
}

func TestP(t *testing.T) {
	got := P("desc", "some text")
	if !strings.Contains(got, `data-host-var="desc"`) {
		t.Fatalf("missing data-host-var: %s", got)
	}
	if !strings.Contains(got, ">some text</p>") {
		t.Fatalf("missing value: %s", got)
	}
}

func TestTag(t *testing.T) {
	got := Tag("em", "count", 42)
	if !strings.Contains(got, `data-host-var="count"`) {
		t.Fatalf("missing data-host-var: %s", got)
	}
	if !strings.Contains(got, ">42</em>") {
		t.Fatalf("missing value: %s", got)
	}
}

// Helper values are HTML-escaped by default so user-derived data cannot
// inject markup through the initial snapshot.
func TestHelpersEscapeValues(t *testing.T) {
	got := Span("msg", `<img src=x onerror=alert(1)>`)
	if !strings.Contains(got, "&lt;img src=x onerror=alert(1)&gt;") {
		t.Fatalf("value not escaped: %s", got)
	}
	if strings.Contains(got, "><img") {
		t.Fatalf("markup injected: %s", got)
	}
	// The expectation hash still covers the unescaped value: the client
	// hashes the element's text content, which the browser unescapes.
	if !strings.Contains(got, encodeExpected(`<img src=x onerror=alert(1)>`)) {
		t.Fatalf("expected hash over raw value: %s", got)
	}
}

// RawTag is the explicit trust API: the value passes through unescaped.
func TestRawTag(t *testing.T) {
	got := RawTag("div", "content", `<b>ok</b>`)
	if !strings.Contains(got, "><b>ok</b></div>") {
		t.Fatalf("raw value escaped: %s", got)
	}
}

func TestRaw(t *testing.T) {
	html := `<div class="custom">foo</div>`
	if got := Raw(html); got != html {
		t.Fatalf("Raw should passthrough: got %q", got)
	}
}

func TestJoin(t *testing.T) {
	got := Join(Span("a", 1), Div("b", 2))
	if !strings.Contains(got, `data-host-var="a"`) {
		t.Fatalf("missing first var: %s", got)
	}
	if !strings.Contains(got, `data-host-var="b"`) {
		t.Fatalf("missing second var: %s", got)
	}
}

func TestExpectedSha1(t *testing.T) {
	got := Span("x", "test")
	expected := encodeExpected("test")
	if !strings.Contains(got, expected) {
		t.Fatalf("expected hash %q not found in %s", expected, got)
	}
}

func TestEncodeExpected(t *testing.T) {
	got := encodeExpected("hello")
	if !strings.HasPrefix(got, "sha1:") {
		t.Fatalf("expected sha1 prefix: %s", got)
	}
}
