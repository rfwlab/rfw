package state

import (
	"fmt"
	"sync"
	"testing"
)

// Stores are advertised for goroutine-driven feeds (dashboards, tickers), so
// concurrent Set/Get/OnChange/Snapshot/history must be data-race free under
// go test -race.
func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore("racestore", WithModule("race"), WithHistory(8))
	s.RegisterComputed(NewComputed("double", []string{"n"}, func(m map[string]any) any {
		if v, ok := m["n"].(int); ok {
			return v * 2
		}
		return 0
	}))
	unwatch := s.RegisterWatcher(NewWatcher([]string{"n"}, func(m map[string]any) {
		_ = m["n"]
	}))
	defer unwatch()

	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				s.Set("n", i)
				s.Set(fmt.Sprintf("k%d", g), i)
				_ = s.Get("n")
				_ = s.Snapshot()
				if i%50 == 0 {
					unsub := s.OnChange("n", func(any) {})
					unsub()
					s.Undo()
					s.Redo()
				}
			}
		}(g)
	}
	wg.Wait()
}

// Signals may be fed from background goroutines while effects and readers run
// elsewhere; Get/Set/OnChange must be data-race free under go test -race.
func TestSignalConcurrentAccess(t *testing.T) {
	sig := NewSignal(0)
	stop := Effect(func() func() {
		_ = sig.Get()
		return nil
	})
	defer stop()
	sub := sig.OnChange(func(int) {})
	defer sub.Stop()

	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				sig.Set(i)
				_ = sig.Get()
				_ = sig.Read()
				sig.SetFromHost(float64(i))
			}
		}()
	}
	wg.Wait()
}
