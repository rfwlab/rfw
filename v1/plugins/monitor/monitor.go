//go:build js && wasm

package monitor

import (
	"encoding/json"
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/events"
)

// Plugin observes DOM mutations and intersections and exposes channels for monitoring.
type Plugin struct {
	MutationSelector     string
	IntersectionSelector string
	IntersectionOpts     jst.Value

	Mutations     chan jst.Value
	Intersections chan jst.Value
}

// New creates a new monitoring plugin.
func New(mSel, iSel string, opts jst.Value) *Plugin {
	return &Plugin{
		MutationSelector:     mSel,
		IntersectionSelector: iSel,
		IntersectionOpts:     opts,
		Mutations:            make(chan jst.Value),
		Intersections:        make(chan jst.Value),
	}
}

func (p *Plugin) Build(json.RawMessage) error { return nil }

// Install hooks into the application and start observers.
func (p *Plugin) Install(a *core.App) {
	if p.MutationSelector != "" {
		ch, _ := events.ObserveMutations(p.MutationSelector)
		go func() {
			for v := range ch {
				p.Mutations <- v
			}
		}()
	}
	if p.IntersectionSelector != "" {
		ch, _ := events.ObserveIntersections(p.IntersectionSelector, p.IntersectionOpts)
		go func() {
			for v := range ch {
				p.Intersections <- v
			}
		}()
	}
}
