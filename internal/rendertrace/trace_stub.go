//go:build !js || !wasm

package rendertrace

import (
	"sync/atomic"
	"time"
)

var (
	batchSeq  atomic.Uint64
	renderSeq atomic.Uint64
)

// Enabled reports that browser tracing is unavailable on native targets.
func Enabled() bool { return false }

// SetEnabledForTest returns a no-op restore function on native targets.
func SetEnabledForTest(bool) func() { return func() {} }

// NextBatchID allocates a native stub batch identifier.
func NextBatchID() uint64 { return batchSeq.Add(1) }

// NextRenderID allocates a native stub render identifier.
func NextRenderID() uint64 { return renderSeq.Add(1) }

// NowMS returns a wall-clock timestamp for native callers.
func NowMS() float64 { return float64(time.Now().UnixNano()) / 1e6 }

// Emit is a no-op because native targets have no browser event transport.
func Emit(Record) {}
