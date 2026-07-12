//go:build js && wasm

package routeranalytics

import (
	"sync"

	"github.com/rfwlab/rfw/v2/hostclient"
)

type wasmPrefetcher struct {
	mu        sync.Mutex
	tick      int64
	requested map[string]struct{}
	channel   string
}

func newPrefetcher(channel string) prefetcher {
	return &wasmPrefetcher{
		channel:   channel,
		requested: make(map[string]struct{}),
	}
}

func (p *wasmPrefetcher) request(resources []string) {
	if len(resources) == 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.requested == nil {
		p.requested = make(map[string]struct{})
	}
	filtered := make([]string, 0, len(resources))
	for _, r := range resources {
		if r == "" {
			continue
		}
		if _, ok := p.requested[r]; ok {
			continue
		}
		p.requested[r] = struct{}{}
		filtered = append(filtered, r)
	}
	if len(filtered) == 0 {
		return
	}
	p.tick++
	// Same wire shape a netcode client would produce: hints are one-way, so
	// the full client (snapshots, interpolation) is not needed here.
	hostclient.Send(p.channel, map[string]any{
		"tick":     p.tick,
		"commands": []any{map[string]any{"resources": filtered}},
	})
}

func (p *wasmPrefetcher) reset() {
	p.mu.Lock()
	p.requested = make(map[string]struct{})
	p.tick = 0
	p.mu.Unlock()
}
