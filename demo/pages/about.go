//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/composition"
	t "github.com/rfwlab/rfw/v2/types"
)

type AboutPage struct{}

func NewAboutPage() *t.View {
	v, err := composition.New(&AboutPage{})
	if err != nil {
		panic(err)
	}
	return v
}