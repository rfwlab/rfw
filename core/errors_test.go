//go:build js && wasm

package core

import (
	"strings"
	"testing"
)

type errPipeComponent struct{ *HTMLComponent }

func newErrPipeComponent() *errPipeComponent {
	c := &errPipeComponent{HTMLComponent: NewHTMLComponent("Boom", []byte("<root><div></div></root>"), nil)}
	c.SetComponent(c)
	c.Init(nil)
	return c
}

func (c *errPipeComponent) Render() string { panic("boom") }

func TestReportErrorFansOutToSinks(t *testing.T) {
	var got []string
	stop := OnError(func(err any, ctx string) {
		got = append(got, ctx)
	})
	defer stop()

	TryRender(newErrPipeComponent())
	if len(got) != 1 || !strings.HasPrefix(got[0], "Render: Boom") {
		t.Fatalf("expected render report, got %v", got)
	}

	b := NewErrorBoundary(newErrPipeComponent(), "<p>fallback</p>")
	out := b.Render()
	if !strings.Contains(out, "fallback") {
		t.Fatalf("expected fallback html, got %q", out)
	}
	if len(got) != 2 || !strings.HasPrefix(got[1], "Boundary render: Boom") {
		t.Fatalf("expected boundary report, got %v", got)
	}

	stop()
	TryRender(newErrPipeComponent())
	if len(got) != 2 {
		t.Fatalf("sink should be removed, got %v", got)
	}
}
