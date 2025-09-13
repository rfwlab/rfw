//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/plugins/shortcut"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/shortcut_component.rtml
var shortcutComponentTpl []byte

// NewShortcutComponent increments a counter when Ctrl+K is pressed.
func NewShortcutComponent() *ShortcutComponent {
	c := &ShortcutComponent{}
	c.count = state.NewSignal(0)
	c.HTMLComponent = core.NewComponentWith(
		"ShortcutComponent",
		shortcutComponentTpl,
		map[string]any{"count": c.count},
		c,
	)
	shortcut.Bind("shift+k", func() { c.count.Set(c.count.Get() + 1) })
	return c
}

type ShortcutComponent struct {
	*core.HTMLComponent
	count *state.Signal[int]
}
