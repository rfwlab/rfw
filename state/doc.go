// Package state provides stores, signals, computed values and watchers.
//
// Concurrency contract: Store and Signal are safe for concurrent use, so
// background goroutines (tickers, feeds, host pushes) may call Set/Get
// directly. Listeners, watchers and dependent effects run synchronously on
// the goroutine that performed the Set, outside internal locks, so they may
// call back into the store or signal. Two things are deliberately not
// parallel-safe: computed functions run under the store lock and must only
// read the state map they receive, and signal dependency tracking (Effect)
// assumes effects never execute in parallel; effects are registered on the
// render goroutine and re-run on whichever goroutine calls Set.
package state
