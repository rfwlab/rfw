//go:build js && wasm

package routeranalytics

import (
	"sync"

	"github.com/rfwlab/rfw/v1/netcode"
)

type prefetchSnapshot struct {
	Resources []string
}

type wasmPrefetcher struct {
	mu        sync.Mutex
	client    *netcode.Client[prefetchSnapshot]
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
	if p.client == nil {
		p.client = netcode.NewClient[prefetchSnapshot](p.channel, decodePrefetchState, interpPrefetchState)
	}
	p.tick++
	p.client.Enqueue(map[string]any{"resources": filtered})
	p.client.Flush(p.tick)
}

func (p *wasmPrefetcher) reset() {
	p.mu.Lock()
	p.requested = make(map[string]struct{})
	p.tick = 0
	p.client = nil
	p.mu.Unlock()
}

func decodePrefetchState(m map[string]any) prefetchSnapshot {
	raw, ok := m["resources"].([]any)
	if !ok {
		return prefetchSnapshot{}
	}
	out := prefetchSnapshot{Resources: make([]string, 0, len(raw))}
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out.Resources = append(out.Resources, s)
		}
	}
	return out
}

func interpPrefetchState(_ prefetchSnapshot, next prefetchSnapshot, _ float64) prefetchSnapshot {
	return next
}
