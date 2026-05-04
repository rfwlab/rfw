//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/composition"
	t "github.com/rfwlab/rfw/v2/types"
)

type HomePage struct {
	Count  t.Int
	Factor t.Int
}

func (h *HomePage) Inc() { h.Count.Set(h.Count.Get() + 1) }
func (h *HomePage) Dec() { h.Count.Set(h.Count.Get() - 1) }

func (h *HomePage) OnMount() {
	h.Count.Set(0)
	h.Factor.Set(2)
}

func NewHomePage() *t.View {
	return composition.New(&HomePage{})
}