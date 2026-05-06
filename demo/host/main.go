package main

import (
	"log"
	"sync/atomic"

	"github.com/rfwlab/rfw/v2/host"
)

// VisitComponent is a struct-based host component.
// It implements host.Component for clean, type-safe server logic.
type VisitComponent struct {
	visits atomic.Int64
}

func (v *VisitComponent) Name() string { return "Visit" }

func (v *VisitComponent) Serve(s *host.Session, p map[string]any) any {
	count := v.visits.Add(1)
	log.Printf("host: Visit handler called, count=%d, payload=%v", count, p)
	return map[string]any{
		"Visit":    count,
		"greeting": "Hello from SSC!",
	}
}

func main() {
	host.RegisterComponent(&VisitComponent{})
	log.Println("SSC host server starting")
	if err := host.StartAuto(); err != nil {
		log.Fatal(err)
	}
}