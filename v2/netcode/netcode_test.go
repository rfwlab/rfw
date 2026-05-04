package netcode

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

type testState struct {
	X float64 `json:"x"`
}

func decodeState(m map[string]any) testState {
	x, _ := m["x"].(float64)
	return testState{X: x}
}

func lerp(a, b testState, alpha float64) testState {
	return testState{X: a.X + (b.X-a.X)*alpha}
}

// Simulates latency between client and server and verifies interpolation.
func TestSync(t *testing.T) {
	cmdCh := make(chan map[string]any, 1)
	snapCh := make(chan map[string]any, 1)

	send := func(_ string, payload any) {
		p := payload.(map[string]any)
		go func() {
			time.Sleep(25 * time.Millisecond)
			cmdCh <- p
		}()
	}
	register := func(_ string, h func(map[string]any)) {
		go func() {
			for p := range snapCh {
				h(p)
			}
		}()
	}

	c := newClient("Game", decodeState, lerp, send, register)
	srv := NewServer("Game", testState{}, func(s *testState, cmd any) {
		m := cmd.(map[string]any)
		if dx, ok := m["dx"].(float64); ok {
			s.X += dx
		}
	})

	go func() {
		for payload := range cmdCh {
			if cmds, ok := payload["commands"].([]any); ok {
				for _, cmd := range cmds {
					srv.apply(&srv.state, cmd)
				}
			}
			b, _ := json.Marshal(srv.Snapshot())
			var m map[string]any
			_ = json.Unmarshal(b, &m)
			snap := map[string]any{
				"tick":  payload["tick"],
				"state": m,
			}
			time.Sleep(25 * time.Millisecond)
			snapCh <- snap
		}
	}()

	c.Enqueue(map[string]any{"dx": 5.0})
	c.Flush(100)
	time.Sleep(100 * time.Millisecond)

	c.Enqueue(map[string]any{"dx": 5.0})
	c.Flush(200)
	time.Sleep(100 * time.Millisecond)

	mid := c.State(150)
	if math.Abs(mid.X-7.5) > 0.1 {
		t.Fatalf("expected ~7.5 got %v", mid.X)
	}
}

func TestServerUpdate(t *testing.T) {
	srv := NewServer("Game", testState{}, func(s *testState, cmd any) {
		if dx, ok := cmd.(float64); ok {
			s.X += dx
		}
	})

	srv.Update(func(s *testState) {
		s.X = 5
	})
	if srv.Snapshot().X != 5 {
		t.Fatalf("expected update to set X to 5")
	}

	srv.Update(func(s *testState) {
		s.X *= 2
	})
	if srv.Snapshot().X != 10 {
		t.Fatalf("expected chained updates to run under lock")
	}

	srv.apply(&srv.state, 1.0)
	if srv.Snapshot().X != 11 {
		t.Fatalf("expected command application to still work, got %v", srv.Snapshot().X)
	}
}
