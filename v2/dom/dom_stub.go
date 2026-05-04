//go:build !js || !wasm

// Package dom provides no-op stubs for non-JS/WASM builds.
package dom

// StoreBindingHook is a stubbed callback for non-wasm builds.
var StoreBindingHook func(componentID, module, store, key string)

// BindStoreInputsForComponent is a no-op outside wasm builds.
func BindStoreInputsForComponent(string, any) {}

// BindStoreInputs is a no-op outside wasm builds.
func BindStoreInputs(any) {}

// SnapshotComponentSignals is a stub returning nil outside wasm builds.
func SnapshotComponentSignals(string) map[string]any { return nil }
