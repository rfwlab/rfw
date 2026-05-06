//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/composition"
	t "github.com/rfwlab/rfw/v2/types"
)

type ContactPage struct{}

func NewContactPage() *t.View {
	v, err := composition.New(&ContactPage{})
	if err != nil {
		panic(err)
	}
	return v
}