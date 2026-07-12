//go:build js && wasm

package components

import (
	"fmt"
	"math/rand/v2"
	"time"
)

var paused bool

func togglePause() {
	paused = !paused
	if paused {
		metrics.Set("status", "paused")
	} else {
		metrics.Set("status", "live")
	}
}

func startFeed(stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		requests := 0
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if paused {
					continue
				}
				cpu := 20 + rand.Float64()*60
				mem := 512 + rand.IntN(1024)
				requests += rand.IntN(50)

				metrics.Set("cpu", fmt.Sprintf("%.1f", cpu))
				metrics.Set("mem", fmt.Sprintf("%d", mem))
				metrics.Set("requests", fmt.Sprintf("%d", requests))

				events, _ := metrics.Get("events").([]any)
				entry := map[string]any{
					"time": time.Now().Format("15:04:05"),
					"msg":  fmt.Sprintf("cpu sample %.1f%%", cpu),
				}
				events = append([]any{entry}, events...)
				if len(events) > 8 {
					events = events[:8]
				}
				metrics.Set("events", events)
			}
		}
	}()
}
