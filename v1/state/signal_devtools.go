//go:build devtools

package state

import "sync"

var (
	sigMu     sync.Mutex
	sigValues = map[int]any{}
	nextSigID int
)

func registerSignal(v any) int {
	sigMu.Lock()
	id := nextSigID
	nextSigID++
	sigValues[id] = v
	sigMu.Unlock()
	return id
}

func updateSignal(id int, v any) {
	sigMu.Lock()
	sigValues[id] = v
	sigMu.Unlock()
	if SignalHook != nil {
		SignalHook(id, v)
	}
}

// SnapshotSignals returns a copy of all tracked signals.
func SnapshotSignals() map[int]any {
	sigMu.Lock()
	defer sigMu.Unlock()
	snap := make(map[int]any, len(sigValues))
	for k, v := range sigValues {
		snap[k] = v
	}
	return snap
}
