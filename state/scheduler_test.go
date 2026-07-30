package state

import "testing"

func TestBatchRunsDependentEffectOnce(t *testing.T) {
	first := NewSignal(1)
	second := NewSignal(2)
	runs := 0
	stop := Effect(func() func() {
		_ = first.Get() + second.Get()
		runs++
		return nil
	})
	defer stop()

	Batch(func() {
		first.Set(3)
		second.Set(4)
	})

	if runs != 2 {
		t.Fatalf("effect runs = %d", runs)
	}
}

func TestUntrackedReadDoesNotBecomeDependency(t *testing.T) {
	tracked := NewSignal(1)
	ignored := NewSignal(2)
	runs := 0
	stop := Effect(func() func() {
		_ = tracked.Get()
		_ = Untracked(ignored.Get)
		runs++
		return nil
	})
	defer stop()

	ignored.Set(3)
	if runs != 1 {
		t.Fatalf("effect tracked untracked signal: runs = %d", runs)
	}
	tracked.Set(2)
	if runs != 2 {
		t.Fatalf("tracked signal did not rerun effect: runs = %d", runs)
	}
}

func TestMemoTracksDependencies(t *testing.T) {
	first := NewSignal(2)
	second := NewSignal(3)
	total := Memo(func() int { return first.Get() + second.Get() })
	defer total.Stop()

	if total.Get() != 5 {
		t.Fatalf("initial memo = %d", total.Get())
	}
	Batch(func() {
		first.Set(4)
		second.Set(5)
	})
	if total.Get() != 9 {
		t.Fatalf("updated memo = %d", total.Get())
	}
}
