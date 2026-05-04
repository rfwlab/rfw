//go:build !js || !wasm

package draw

// NewCanvas is unavailable outside WebAssembly builds.
func NewCanvas(_ any) (Canvas, bool) { return Canvas{}, false }

type noopCanvas struct{}

func (noopCanvas) valid() bool                   { return false }
func (noopCanvas) setSize(width, height float64) {}
func (noopCanvas) drawRect(Rect)                 {}
func (noopCanvas) drawCircle(Circle)             {}
func (noopCanvas) drawLine(Line)                 {}
