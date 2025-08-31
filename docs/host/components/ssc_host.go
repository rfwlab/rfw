package components

import (
	"time"

	"github.com/rfwlab/rfw/v1/host"
)

func RegisterSSCHost() {
	var counter int
	host.Register(host.NewHostComponent("SSCHost", func(_ map[string]any) any {
		return map[string]any{"value": counter}
	}))
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			counter++
			host.Broadcast("SSCHost", map[string]any{"value": counter})
		}
	}()
}
