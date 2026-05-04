package main

import (
	"log"
	"sync/atomic"

	"github.com/rfwlab/rfw/v2/host"
)

var visits atomic.Int64

func main() {
	hc := host.NewHostComponentWithSession("demo-counter", func(s *host.Session, p map[string]any) any {
		count := visits.Add(1)
		return map[string]any{
			"visit":    count,
			"greeting": "Hello from SSC!",
		}
	})
	hc.WithInitSnapshot(func(s *host.Session, p map[string]any) *host.InitSnapshot {
		return &host.InitSnapshot{
			HTML: `<span data-signal="visit">1</span>`,
			Vars: []string{"visit"},
		}
	})
	host.Register(hc)

	log.Println("SSC host server starting")
	if err := host.Start("../build/client"); err != nil {
		log.Fatal(err)
	}
}