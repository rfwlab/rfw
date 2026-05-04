//go:build !js || !wasm

// Package virtual provides no-op stubs for non-JS/WASM builds.
package virtual

// VirtualList is a placeholder that does nothing on non-JS/WASM platforms.
type VirtualList struct{}

// NewVirtualList returns an empty VirtualList placeholder.
func NewVirtualList(_ string, _ int, _ int, _ func(int) string) *VirtualList {
	return &VirtualList{}
}

// Destroy performs no action in non-JS/WASM builds.
func (v *VirtualList) Destroy() {}
