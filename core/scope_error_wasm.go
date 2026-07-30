//go:build js && wasm

package core

func reportScopeError(err any) {
	ReportError(err, "component scope cleanup")
}
