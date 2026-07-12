//go:build js && wasm

package hostclient

import "testing"

func pendingCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(pending)
}

// Repeated identical messages must go through by default: two identical user
// actions within the dedup window (e.g. clicking +1 twice) are intentional.
func TestSendRepeatedMessagesNotDeduped(t *testing.T) {
	before := pendingCount()
	Send("CounterHost", map[string]any{"cmd": "increment"})
	Send("CounterHost", map[string]any{"cmd": "increment"})
	if got := pendingCount() - before; got != 2 {
		t.Fatalf("expected 2 queued messages, got %d", got)
	}
}

// Dedup is opt-in per channel: after EnableSendDedup identical payloads within
// the TTL window are dropped.
func TestSendDedupOptIn(t *testing.T) {
	EnableSendDedup("DedupHost")
	before := pendingCount()
	Send("DedupHost", map[string]any{"cmd": "refresh"})
	Send("DedupHost", map[string]any{"cmd": "refresh"})
	if got := pendingCount() - before; got != 1 {
		t.Fatalf("expected 1 queued message after dedup, got %d", got)
	}
}
