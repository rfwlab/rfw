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