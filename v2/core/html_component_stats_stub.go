//go:build !js || !wasm

package core

import "time"

// ComponentStats is a stub for non-wasm builds.
type ComponentStats struct {
	RenderCount   int
	TotalRender   time.Duration
	LastRender    time.Duration
	AverageRender time.Duration
	Timeline      []ComponentTimelineEntry
}

// ComponentTimelineEntry is a stub for non-wasm builds.
type ComponentTimelineEntry struct {
	Kind      string
	Timestamp time.Time
	Duration  time.Duration
}

// Stats returns zeroed metrics on non-wasm builds.
func (c *HTMLComponent) Stats() ComponentStats { return ComponentStats{} }
