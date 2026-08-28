//go:build js && wasm

package rendertrace

import (
	"sync/atomic"
	"syscall/js"
)

const schemaVersion = 1

var (
	bootEnabled       = traceEnabledAtBoot()
	testOverride int8 = -1
	sequence     atomic.Uint64
	batchSeq     atomic.Uint64
	renderSeq    atomic.Uint64
)

func traceEnabledAtBoot() bool {
	value := js.Global().Get("RFW_RENDER_TRACE")
	if value.Type() == js.TypeBoolean {
		return value.Bool()
	}
	if value.Type() == js.TypeObject {
		enabled := value.Get("enabled")
		return enabled.Type() == js.TypeBoolean && enabled.Bool()
	}
	return false
}

// Enabled reports whether tracing was enabled before WASM startup. The test
// override exists inside the module's internal package so browser tests can
// exercise both paths without exposing a production API.
func Enabled() bool {
	if testOverride >= 0 {
		return testOverride == 1
	}
	return bootEnabled
}

// SetEnabledForTest overrides the boot flag and returns a restoring function.
func SetEnabledForTest(enabled bool) func() {
	previous := testOverride
	if enabled {
		testOverride = 1
	} else {
		testOverride = 0
	}
	return func() { testOverride = previous }
}

// NextBatchID allocates a monotonic render batch identifier.
func NextBatchID() uint64 { return batchSeq.Add(1) }

// NextRenderID allocates a monotonic logical render identifier.
func NextRenderID() uint64 { return renderSeq.Add(1) }

// NowMS returns the browser's monotonic clock when it is available.
func NowMS() float64 {
	performance := js.Global().Get("performance")
	if performance.Type() == js.TypeObject && performance.Get("now").Type() == js.TypeFunction {
		return performance.Call("now").Float()
	}
	return js.Global().Get("Date").Call("now").Float()
}

// Emit dispatches one browser-readable trace record. Tracing must never affect
// application rendering, so unsupported browser APIs or consumer errors are
// contained here.
func Emit(record Record) {
	if !Enabled() {
		return
	}
	defer func() { _ = recover() }()

	detail := map[string]any{
		"schemaVersion": schemaVersion,
		"sequence":      sequence.Add(1),
		"timestampMs":   NowMS(),
		"event":         record.Event,
		"batchId":       record.BatchID,
		"renderId":      record.RenderID,
		"componentId":   record.ComponentID,
		"componentName": record.ComponentName,
		"depth":         record.Depth,
	}
	if record.ParentComponentID != "" {
		detail["parentComponentId"] = record.ParentComponentID
	}
	if record.Cause.Kind != "" {
		detail["cause"] = causeDetail(record.Cause)
	}
	if len(record.Causes) > 0 {
		causes := make([]any, 0, len(record.Causes))
		for _, cause := range record.Causes {
			causes = append(causes, causeDetail(cause))
		}
		detail["causes"] = causes
	}
	if record.QueueDepth > 0 {
		detail["queueDepth"] = record.QueueDepth
	}
	if record.CoalescedCount > 0 || record.Event == "scheduled" || record.Event == "coalesced" || record.Event == "started" || record.Event == "committed" {
		detail["coalescedCount"] = record.CoalescedCount
	}
	if record.Event == "committed" || record.Event == "failed" {
		detail["templateMs"] = record.TemplateMS
		detail["domMs"] = record.DOMMS
		detail["totalMs"] = record.TotalMS
	}
	if record.Outcome != "" {
		detail["outcome"] = record.Outcome
	}
	if record.Reason != "" {
		detail["reason"] = record.Reason
	}
	if record.SupersededBy != 0 {
		detail["supersededByRenderId"] = record.SupersededBy
	}

	constructor := js.Global().Get("CustomEvent")
	dispatch := js.Global().Get("dispatchEvent")
	if constructor.Type() != js.TypeFunction || dispatch.Type() != js.TypeFunction {
		return
	}
	options := map[string]any{"detail": detail}
	event := constructor.New("rfw:render-trace", options)
	js.Global().Call("dispatchEvent", event)
}

func causeDetail(cause Cause) map[string]any {
	cause = NormalizeCause(cause)
	detail := map[string]any{"kind": cause.Kind}
	if cause.Module != "" {
		detail["module"] = cause.Module
	}
	if cause.Store != "" {
		detail["store"] = cause.Store
	}
	if cause.Key != "" {
		detail["key"] = cause.Key
	}
	if cause.Signal != "" {
		detail["signal"] = cause.Signal
	}
	return detail
}
