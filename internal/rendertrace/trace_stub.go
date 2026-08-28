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

func Enabled() bool                 { return false }
func SetEnabledForTest(bool) func() { return func() {} }
func NextBatchID() uint64           { return batchSeq.Add(1) }
func NextRenderID() uint64          { return renderSeq.Add(1) }
func NowMS() float64                { return float64(time.Now().UnixNano()) / 1e6 }
func Emit(Record)                   {}
