//go:build js && wasm

package js

// SetTimeout schedules fn to run after delayMs milliseconds and returns the
// timer id. It wraps window.setTimeout so callers need not touch the global.
func SetTimeout(fn func(), delayMs int) Value {
	var cb Func
	cb = FuncOf(func(_ Value, _ []Value) any {
		fn()
		cb.Release()
		return nil
	})
	return Call("setTimeout", cb, delayMs)
}
