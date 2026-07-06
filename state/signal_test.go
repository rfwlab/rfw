package state

import "testing"

func TestSignalEffect(t *testing.T) {
	a := NewSignal(0)
	b := NewSignal(0)

	var runs int
	stop := Effect(func() func() {
		_ = a.Get()
		runs++
		return nil
	})
	defer stop()

	b.Set(1)
	if runs != 1 {
		t.Fatalf("effect ran on unrelated signal change")
	}
	a.Set(1)
	if runs != 2 {
		t.Fatalf("effect did not rerun on dependent signal change")
	}
}

func TestEffectCleanup(t *testing.T) {
	s := NewSignal(0)
	var cleans int
	stop := Effect(func() func() {
		_ = s.Get()
		return func() { cleans++ }
	})

	s.Set(1)
	if cleans != 1 {
		t.Fatalf("cleanup not called before rerun, got %d", cleans)
	}
	stop()
	if cleans != 2 {
		t.Fatalf("cleanup not called on stop, got %d", cleans)
	}
}

func TestSetNotifiesAllSubscribers(t *testing.T) {
	s := NewSignal(0)

	var val1, val2 int
	stop1 := Effect(func() func() {
		val1 = s.Get()
		return nil
	})
	defer stop1()

	stop2 := Effect(func() func() {
		val2 = s.Get() * 2
		return nil
	})
	defer stop2()

	if val1 != 0 || val2 != 0 {
		t.Fatalf("initial: expected 0,0 got %d,%d", val1, val2)
	}

	s.Set(5)
	if val1 != 5 {
		t.Fatalf("effect1 after Set(5): expected 5, got %d", val1)
	}
	if val2 != 10 {
		t.Fatalf("effect2 after Set(5): expected 10, got %d", val2)
	}
}

func TestExprEffectWithMultipleSignals(t *testing.T) {
	count := NewSignal(0)
	factor := NewSignal(0)

	var result int
	stop := Effect(func() func() {
		result = count.Get() * factor.Get()
		return nil
	})
	defer stop()

	if result != 0 {
		t.Fatalf("initial: expected 0, got %d", result)
	}

	factor.Set(2)
	if result != 0 {
		t.Fatalf("after factor=2: expected 0, got %d", result)
	}

	count.Set(1)
	if result != 2 {
		t.Fatalf("after count=1: expected 2, got %d", result)
	}

	count.Set(3)
	if result != 6 {
		t.Fatalf("after count=3: expected 6, got %d", result)
	}

	factor.Set(3)
	if result != 9 {
		t.Fatalf("after factor=3: expected 9, got %d", result)
	}
}

func TestOnChangeBasic(t *testing.T) {
	s := NewSignal("hello")
	var received []string

	sub := s.OnChange(func(v string) {
		received = append(received, v)
	})

	s.Set("world")
	s.Set("!")

	if len(received) != 2 || received[0] != "world" || received[1] != "!" {
		t.Fatalf("expected [world !], got %v", received)
	}

	sub.Stop()
	s.Set("after-stop")
	if len(received) != 2 {
		t.Fatalf("callback fired after Stop: %v", received)
	}
}

func TestOnChangeMultipleListeners(t *testing.T) {
	s := NewSignal(0)
	var sum int

	sub1 := s.OnChange(func(v int) { sum += v })
	sub2 := s.OnChange(func(v int) { sum += v * 10 })

	s.Set(1)
	if sum != 11 {
		t.Fatalf("expected 11, got %d", sum)
	}

	sub1.Stop()
	s.Set(2)
	if sum != 31 {
		t.Fatalf("expected 31, got %d", sum)
	}

	sub2.Stop()
	s.Set(3)
	if sum != 31 {
		t.Fatalf("expected 31 after all stopped, got %d", sum)
	}
}

func TestOnChangeStopsAreIdempotent(t *testing.T) {
	s := NewSignal(0)
	calls := 0

	sub := s.OnChange(func(v int) { calls++ })

	s.Set(1)
	if calls != 1 {
		t.Fatalf("expected 1, got %d", calls)
	}

	sub.Stop()
	sub.Stop()
	sub.Stop()

	s.Set(2)
	if calls != 1 {
		t.Fatalf("callback fired after multiple stops: %d", calls)
	}
}

func TestChannelLazyCreation(t *testing.T) {
	s := NewSignal(0)

	s.onChangeMu.Lock()
	hasCh := s.chCreated
	s.onChangeMu.Unlock()
	if hasCh {
		t.Fatal("channel should not exist before Channel() call")
	}

	ch := s.Channel()

	s.onChangeMu.Lock()
	hasCh = s.chCreated
	s.onChangeMu.Unlock()
	if !hasCh {
		t.Fatal("channel should exist after Channel() call")
	}
	_ = ch
}

func TestChannelReceivesValues(t *testing.T) {
	s := NewSignal(0)
	ch := s.Channel()

	s.Set(42)

	select {
	case v := <-ch:
		if v != 42 {
			t.Fatalf("expected 42, got %d", v)
		}
	default:
		t.Fatal("channel should have received value")
	}
}

func TestChannelClosesWhenAllListenersRemoved(t *testing.T) {
	s := NewSignal(0)

	s.onChangeMu.Lock()
	hasCh := s.chCreated
	s.onChangeMu.Unlock()
	if hasCh {
		t.Fatal("channel should not exist before Channel() or OnChange()")
	}
}

func TestOnChangeDoesNotFireAfterStop(t *testing.T) {
	s := NewSignal(10)
	calls := 0

	sub := s.OnChange(func(v int) { calls++ })

	s.Set(20)
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	sub.Stop()

	s.Set(30)
	s.Set(40)
	if calls != 1 {
		t.Fatalf("expected calls to stay 1 after Stop, got %d", calls)
	}
}