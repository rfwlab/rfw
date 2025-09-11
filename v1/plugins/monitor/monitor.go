//go:build js && wasm

package monitor

import (
	"encoding/json"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
)

// Plugin observes DOM mutations and intersections and exposes channels for monitoring.
type Plugin struct {
	MutationSelector     string
	IntersectionSelector string
	IntersectionOpts     js.Value

	Mutations     chan js.Value
	Intersections chan js.Value
}

// New creates a new monitoring plugin.
func New(mSel, iSel string, opts js.Value) *Plugin {
	return &Plugin{
		MutationSelector:     mSel,
		IntersectionSelector: iSel,
		IntersectionOpts:     opts,
		Mutations:            make(chan js.Value),
		Intersections:        make(chan js.Value),
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
