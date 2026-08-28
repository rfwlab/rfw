//go:build js && wasm

package rendertrace

import (
	"syscall/js"
	"testing"
)

func TestTraceEnabledAtBootReadsBooleanAndObjectFlags(t *testing.T) {
	global := js.Global()
	previous := global.Get("RFW_RENDER_TRACE")
	defer func() {
		if previous.Type() == js.TypeUndefined {
			global.Get("Reflect").Call("deleteProperty", global, "RFW_RENDER_TRACE")
			return
		}
		global.Set("RFW_RENDER_TRACE", previous)
	}()

	global.Set("RFW_RENDER_TRACE", true)
	if !traceEnabledAtBoot() {
		t.Fatal("boolean true did not enable tracing")
	}
	global.Set("RFW_RENDER_TRACE", map[string]any{"enabled": true})
	if !traceEnabledAtBoot() {
		t.Fatal("object enabled flag did not enable tracing")
	}
	global.Set("RFW_RENDER_TRACE", false)
	if traceEnabledAtBoot() {
		t.Fatal("boolean false enabled tracing")
	}
}
