//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/composition"
	t "github.com/rfwlab/rfw/v2/types"
)

type HomePage struct {
	Count  t.Int
	Factor t.Int
	Visit  t.HInt
}

func (h *HomePage) Inc()  { h.Count.Set(h.Count.Get() + 1) }
func (h *HomePage) Dec()  { h.Count.Set(h.Count.Get() - 1) }

func (h *HomePage) OnMount() {
	(&h.Count).Set(2)
	(&h.Factor).Set(2)
}

func NewHomePage() *t.View {
	v, err := composition.New(&HomePage{})
	if err != nil {
		panic(err)
	}
	return v
}