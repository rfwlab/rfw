//go:build js && wasm

package core

import (
	"fmt"
	"testing"

	"github.com/rfwlab/rfw/v2/state"
)

// benchIssues builds a collection shaped like the issue board that motivated
// this benchmark: many rows, several fields each.
func benchIssues(n int) []any {
	items := make([]any, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, map[string]any{
			"id":       fmt.Sprintf("%d", i),
			"key":      fmt.Sprintf("DAT-%d", 1000+i),
			"title":    fmt.Sprintf("Issue number %d needs a decision", i),
			"status":   []string{"open", "in progress", "review", "done"}[i%4],
			"assignee": []string{"mirko", "alex", "pietro"}[i%3],
			"priority": []string{"low", "normal", "high"}[i%3],
			"due":      fmt.Sprintf("2026-08-%02d", i%28+1),
		})
	}
	return items
}

const benchBoardTemplate = `<root>
<div class="board flex flex-col gap-2">
@if:store:app.bench.ready=="yes"
<div class="toolbar flex items-center gap-2 border-b border-border">@store:app.bench.title</div>
<div class="count text-xs text-muted">@expr:total</div>
@endif
<div class="col flex flex-col gap-1" data-col="open">
@for:issue in store:app.bench.issues
<article class="card border border-border bg-surface p-2" data-id="@prop:issue.id" @on:click:openIssue>
<div class="flex items-center gap-2">
<span class="key text-xs font-mono">@prop:issue.key</span>
<span class="due text-xs text-muted">@prop:issue.due</span>
</div>
<h3 class="title text-sm">@prop:issue.title</h3>
<div class="meta flex gap-1 text-xs">
<span class="status">@prop:issue.status</span>
<span class="assignee">@prop:issue.assignee</span>
<span class="priority">@prop:issue.priority</span>
</div>
</article>
@endfor
</div>
</div>
</root>`

func newBenchBoard(b *testing.B, rows int) *HTMLComponent {
	b.Helper()
	st := state.NewStore("bench", state.WithModule("app"))
	st.Set("ready", "yes")
	st.Set("title", "ISSUES")
	st.Set("issues", benchIssues(rows))
	c := NewHTMLComponent("BenchBoard", []byte(benchBoardTemplate), map[string]any{"total": rows})
	c.Init(nil)
	return c
}

// BenchmarkRenderBoard118 renders the whole component the way a store change
// does: cache dropped, every directive pass re-run over the expanded markup.
func BenchmarkRenderBoard118(b *testing.B) {
	c := newBenchBoard(b, 118)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if out := c.RenderFresh(); out == "" {
			b.Fatal("empty render")
		}
	}
}

// BenchmarkForExpansion118 isolates the loop expansion from the passes that
// run over its output.
func BenchmarkForExpansion118(b *testing.B) {
	c := newBenchBoard(b, 118)
	tpl := c.Template
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if out := replaceForPlaceholders(tpl, c); out == "" {
			b.Fatal("empty expansion")
		}
	}
}

// BenchmarkRenderBoardStatic renders a board with no directives left in the
// rows, which is what the post-expansion passes actually scan.
func BenchmarkRenderBoardStatic(b *testing.B) {
	c := newBenchBoard(b, 118)
	expanded := replaceForPlaceholders(c.Template, c)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := replaceStorePlaceholders(expanded, c)
		out = replaceSignalPlaceholders(out, c)
		out = replaceExprInClassAttr(out, c)
		out = replaceExprPlaceholders(out, c)
		out = replacePropPlaceholders(out, c)
		out = replacePluginPlaceholders(out)
		out = replaceConditionals(out, c)
		out = replaceEventHandlers(out)
		out = replaceRtIsAttributes(out, c)
		out = replaceConstructors(out)
		out = minifyInline(out)
		if out == "" {
			b.Fatal("empty render")
		}
	}
}
