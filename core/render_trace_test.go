//go:build js && wasm

package core

import (
	"syscall/js"
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/internal/rendertrace"
	"github.com/rfwlab/rfw/v2/state"
)

type renderTraceSnapshot struct {
	event          string
	batchID        int
	renderID       int
	componentID    string
	causeKind      string
	module         string
	store          string
	key            string
	coalescedCount int
	hasTimings     bool
	reason         string
	templateMS     float64
	domMS          float64
}

func captureRenderTrace(t *testing.T) (*[]renderTraceSnapshot, func()) {
	t.Helper()
	records := []renderTraceSnapshot{}
	callback := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		detail := args[0].Get("detail")
		cause := detail.Get("cause")
		record := renderTraceSnapshot{
			event:       detail.Get("event").String(),
			batchID:     detail.Get("batchId").Int(),
			renderID:    detail.Get("renderId").Int(),
			componentID: detail.Get("componentId").String(),
			hasTimings:  detail.Call("hasOwnProperty", "templateMs").Bool() && detail.Call("hasOwnProperty", "domMs").Bool(),
		}
		if cause.Type() == js.TypeObject {
			record.causeKind = cause.Get("kind").String()
			if value := cause.Get("module"); value.Type() == js.TypeString {
				record.module = value.String()
			}
			if value := cause.Get("store"); value.Type() == js.TypeString {
				record.store = value.String()
			}
			if value := cause.Get("key"); value.Type() == js.TypeString {
				record.key = value.String()
			}
		}
		if value := detail.Get("coalescedCount"); value.Type() == js.TypeNumber {
			record.coalescedCount = value.Int()
		}
		if value := detail.Get("reason"); value.Type() == js.TypeString {
			record.reason = value.String()
		}
		if value := detail.Get("templateMs"); value.Type() == js.TypeNumber {
			record.templateMS = value.Float()
		}
		if value := detail.Get("domMs"); value.Type() == js.TypeNumber {
			record.domMS = value.Float()
		}
		records = append(records, record)
		return nil
	})
	js.Global().Call("addEventListener", "rfw:render-trace", callback)
	return &records, func() {
		js.Global().Call("removeEventListener", "rfw:render-trace", callback)
		callback.Release()
	}
}

func TestRenderTraceCapturesStoreCoalescing(t *testing.T) {
	restoreTrace := rendertrace.SetEnabledForTest(true)
	defer restoreTrace()
	records, stopCapture := captureRenderTrace(t)
	defer stopCapture()

	store := state.NewStore("trace_scheduler", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "trace_scheduler")
	store.Set("items", []any{map[string]any{"label": "before"}})
	component := mountForComponent(t, "TraceScheduler", []byte(`<root><ul>
@for:item in store:app.trace_scheduler.items
<li>@prop:item.label</li>
@endfor
</ul></root>`))
	defer component.Unmount()

	store.Set("items", []any{map[string]any{"label": "one"}})
	store.Set("items", []any{map[string]any{"label": "two"}})
	waitForRenderFlush()

	var scheduled, coalesced, started, committed *renderTraceSnapshot
	for index := range *records {
		record := &(*records)[index]
		if record.componentID != component.ID {
			continue
		}
		switch record.event {
		case "scheduled":
			scheduled = record
		case "coalesced":
			coalesced = record
		case "started":
			started = record
		case "committed":
			committed = record
		}
	}
	if scheduled == nil || coalesced == nil || started == nil || committed == nil {
		t.Fatalf("missing trace phases: %#v", *records)
	}
	if scheduled.batchID == 0 || scheduled.renderID == 0 || scheduled.batchID != committed.batchID || scheduled.renderID != committed.renderID {
		t.Fatalf("trace identity changed across phases: scheduled=%#v committed=%#v", scheduled, committed)
	}
	if scheduled.causeKind != "store" || scheduled.module != "app" || scheduled.store != "trace_scheduler" || scheduled.key != "items" {
		t.Fatalf("store provenance = %#v", scheduled)
	}
	if coalesced.coalescedCount != 1 || committed.coalescedCount != 1 {
		t.Fatalf("coalescing facts: coalesced=%#v committed=%#v", coalesced, committed)
	}
	if !committed.hasTimings {
		t.Fatalf("terminal trace has no timing fields: %#v", committed)
	}
}

func TestRenderTraceDisabledDispatchesNothing(t *testing.T) {
	restoreTrace := rendertrace.SetEnabledForTest(false)
	defer restoreTrace()
	records, stopCapture := captureRenderTrace(t)
	defer stopCapture()

	done := false
	requestScheduledRender(renderJob{
		id:       "trace-disabled",
		active:   func() bool { return true },
		evaluate: func() string { return "" },
		commit:   func(string) { done = true },
	})
	waitForRenderFlush()
	if !done {
		t.Fatal("disabled tracing changed scheduler execution")
	}
	if len(*records) != 0 {
		t.Fatalf("disabled tracing dispatched %d records", len(*records))
	}
}

func TestRenderTraceReportsCancellationAndUnmount(t *testing.T) {
	restoreTrace := rendertrace.SetEnabledForTest(true)
	defer restoreTrace()
	records, stopCapture := captureRenderTrace(t)
	defer stopCapture()

	store := state.NewStore("trace_unmount", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "trace_unmount")
	store.Set("items", []any{})
	component := mountForComponent(t, "TraceUnmount", []byte(`<root><ul>
@for:item in store:app.trace_unmount.items
<li>@prop:item</li>
@endfor
</ul></root>`))
	store.Set("items", []any{"queued"})
	component.Unmount()
	waitForRenderFlush()

	var cancelled, unmounted bool
	for _, record := range *records {
		if record.componentID != component.ID {
			continue
		}
		cancelled = cancelled || record.event == "cancelled"
		unmounted = unmounted || record.event == "unmounted"
	}
	if !cancelled || !unmounted {
		t.Fatalf("cleanup trace missing: %#v", *records)
	}
}

func TestRenderTraceReportsFailedJob(t *testing.T) {
	restoreTrace := rendertrace.SetEnabledForTest(true)
	defer restoreTrace()
	records, stopCapture := captureRenderTrace(t)
	defer stopCapture()
	stopErrors := OnError(func(any, string) {})
	defer stopErrors()

	requestScheduledRender(renderJob{
		id:     "trace-failure",
		active: func() bool { return true },
		trace:  newRenderJobTrace("TraceFailure", "", nil),
		evaluate: func() string {
			time.Sleep(3 * time.Millisecond)
			panic("trace failure")
		},
		commit: func(string) {},
	})
	waitForRenderFlush()

	for _, record := range *records {
		if record.componentID == "trace-failure" && record.event == "failed" {
			if !record.hasTimings || record.reason != "trace failure" || record.templateMS < 2 {
				t.Fatalf("failed trace = %#v", record)
			}
			return
		}
	}
	t.Fatalf("failed trace missing: %#v", *records)
}
