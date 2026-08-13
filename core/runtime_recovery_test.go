//go:build js && wasm

package core

import (
	"strings"
	"testing"

	js "github.com/rfwlab/rfw/v2/js"
)

func TestSafeJavaScriptCallbackReportsAndRuntimeContinues(t *testing.T) {
	contexts := []string{}
	stop := OnError(func(_ any, context string) { contexts = append(contexts, context) })
	defer stop()

	calls := 0
	callback := js.SafeFuncOf(func(js.Value, []js.Value) any {
		calls++
		if calls == 1 {
			panic("callback failure")
		}
		return calls
	})
	defer callback.Release()

	callback.Invoke()
	result := callback.Invoke()

	if calls != 2 || result.Int() != 2 {
		t.Fatalf("callback did not survive: calls=%d result=%v", calls, result)
	}
	if len(contexts) != 1 || !strings.Contains(contexts[0], "JavaScript callback") {
		t.Fatalf("unexpected recovery contexts: %v", contexts)
	}
}

func TestErrorSinkPanicDoesNotBlockOtherSinks(t *testing.T) {
	stopBroken := OnError(func(any, string) { panic("sink failure") })
	defer stopBroken()
	called := 0
	stopHealthy := OnError(func(any, string) { called++ })
	defer stopHealthy()

	ReportError("original", "test")

	if called != 1 {
		t.Fatalf("healthy error sink calls = %d, want 1", called)
	}
}
