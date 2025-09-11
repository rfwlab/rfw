package components

import (
	"time"

	"github.com/rfwlab/rfw/v1/host"
	"github.com/rfwlab/rfw/v1/netcode"
)

type ncState struct {
	X float64 `json:"x"`
}

// RegisterNetcodeHost sets up a netcode server broadcasting position updates.
func RegisterNetcodeHost() {
	srv := netcode.NewServer("NetcodeHost", ncState{}, func(s *ncState, cmd any) {
		if m, ok := cmd.(map[string]any); ok {
			if dx, ok := m["dx"].(float64); ok {
				s.X += dx
			}
		}
	})
	host.Register(srv.HostComponent())
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		var tick int64
		for range ticker.C {
			tick += 50
			srv.Broadcast(tick)
		}
	}()
}
