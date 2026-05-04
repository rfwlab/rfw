//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/composition"
	t "github.com/rfwlab/rfw/v2/types"
)

type AboutPage struct{}

func NewAboutPage() *t.View {
	return composition.New(&AboutPage{})
}