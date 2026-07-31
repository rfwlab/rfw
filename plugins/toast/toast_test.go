//go:build js && wasm

package toast

import "testing"

func TestPushOptionsNoPlugin(_ *testing.T) {
	// Ensure calling PushOptions without installation is a no-op.
	PushOptions("msg", Options{})
}
