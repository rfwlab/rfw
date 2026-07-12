//go:build js && wasm

package components

import "github.com/rfwlab/rfw/v2/state"

// metrics is registered globally as module "app", store "metrics",
// so templates reference it as @store:app.metrics.<key>.
var metrics = state.NewStore("metrics", state.WithModule("app"))

func seedMetrics() {
	metrics.Set("status", "live")
	metrics.Set("cpu", "0.0")
	metrics.Set("mem", "0")
	metrics.Set("requests", "0")
	metrics.Set("events", []any{})
}
