//go:build js && wasm

package shortcut

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
)

// Plugin listens for keyboard combinations and runs registered handlers.
type Plugin struct {
	bindings map[string]func()
	pressed  map[string]bool
}

var current *Plugin

// New creates a keyboard shortcut plugin instance.
func New() *Plugin {
	return &Plugin{bindings: make(map[string]func()), pressed: make(map[string]bool)}
}

// Build is a no-op.
func (p *Plugin) Build(json.RawMessage) error { return nil }

// Install registers key listeners and activates the plugin.
func (p *Plugin) Install(a *core.App) {
	current = p
	events.OnKeyDown(func(e js.Value) {
		key := strings.ToLower(e.Get("key").String())
		p.pressed[key] = true
		combo := p.combo()
		if fn, ok := p.bindings[combo]; ok {
			fn()
		}
	})
	events.OnKeyUp(func(e js.Value) {
		key := strings.ToLower(e.Get("key").String())
		delete(p.pressed, key)
	})
}

// Bind associates combo (e.g. "control+k") with fn.
func Bind(combo string, fn func()) {
	if current == nil {
		return
	}
	current.bindings[normalize(combo)] = fn
}

func (p *Plugin) combo() string {
	keys := make([]string, 0, len(p.pressed))
	for k := range p.pressed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, "+")
}

func normalize(combo string) string {
	parts := strings.Split(strings.ToLower(combo), "+")
	sort.Strings(parts)
	return strings.Join(parts, "+")
}
