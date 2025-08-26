//go:build js && wasm

package monitor

import (
	"log"

	"github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
	m "github.com/rfwlab/rfw/v1/plugins/monitor"
)

// New creates a monitoring plugin that logs observed events.
func New() core.Plugin {
	p := m.New("body", "img", js.Null())
	go func() {
		for v := range p.Mutations {
			log.Printf("mutation: %v", v)
		}
	}()
	go func() {
		for v := range p.Intersections {
			log.Printf("intersection: %v", v)
		}
	}()
	return p
}
