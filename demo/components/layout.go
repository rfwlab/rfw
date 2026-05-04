//go:build js && wasm

package components

import (
	"github.com/rfwlab/rfw/v2/composition"
	"github.com/rfwlab/rfw/v2/router"
	t "github.com/rfwlab/rfw/v2/types"
)

type Layout struct {
	Content    *t.View
	ActivePath *t.String
}

func NewLayout(content *t.View) *t.View {
	return composition.New(&Layout{
		Content:    content,
		ActivePath: router.ActivePath(),
	})
}