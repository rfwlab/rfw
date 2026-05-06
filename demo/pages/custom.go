//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/composition"
	t "github.com/rfwlab/rfw/v2/types"
)

type CustomPage struct {
	Title t.String
}

func (c *CustomPage) OnUnmount() {
	c.Title.Set("")
}

func NewCustomPage() *t.View {
	v, err := composition.New(&CustomPage{})
	if err != nil {
		panic(err)
	}
	return v
}