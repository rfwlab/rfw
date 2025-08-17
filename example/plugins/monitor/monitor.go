//go:build js && wasm

package monitor

import (
	"log"
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/core"
	m "github.com/rfwlab/rfw/v1/plugins/monitor"
)

// New creates a monitoring plugin that logs observed events.
func New() core.Plugin {
	p := m.New("body", "img", jst.Null())
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
