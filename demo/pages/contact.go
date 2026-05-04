//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/composition"
	t "github.com/rfwlab/rfw/v2/types"
)

type ContactPage struct{}

func NewContactPage() *t.View {
	return composition.New(&ContactPage{})
}