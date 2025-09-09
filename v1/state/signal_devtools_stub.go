//go:build !devtools

package state

func registerSignal(v any) int     { return 0 }
func updateSignal(id int, v any)   {}
func SnapshotSignals() map[int]any { return nil }
