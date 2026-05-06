package main

import (
	"log"
	"sync/atomic"

	"github.com/rfwlab/rfw/v2/host"
)

var visits atomic.Int64

func main() {
	hc := host.NewHostComponentWithSession("Visit", func(s *host.Session, p map[string]any) any {
		count := visits.Add(1)
		log.Printf("host: Visit handler called, count=%d, payload=%v", count, p)
		return map[string]any{
			"Visit":    count,
			"greeting": "Hello from SSC!",
		}
	})
	hc.WithInitSnapshot(func(s *host.Session, p map[string]any) *host.InitSnapshot {
		return &host.InitSnapshot{
			HTML: `<span data-host-var="Visit">1</span>`,
			Vars: []string{"Visit"},
		}
	})
	host.Register(hc)

	log.Println("SSC host server starting")
	if err := host.Start("../build/client"); err != nil {
		log.Fatal(err)
	}
}