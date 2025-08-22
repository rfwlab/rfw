package state

import (
	"fmt"
	"sync"
	"testing"
)

func TestStoreConcurrency(t *testing.T) {
	s := NewStore("race")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(4)
		go func(i int) {
			s.Set("key", i)
			wg.Done()
		}(i)
		go func() {
			_ = s.Get("key")
			wg.Done()
		}()
		go func() {
			c := NewComputed(fmt.Sprintf("comp%d", i), []string{"key"}, func(m map[string]any) any {
				return m["key"]
			})
			s.RegisterComputed(c)
			wg.Done()
		}()
		go func() {
			w := NewWatcher([]string{"key"}, func(map[string]any) {})
			s.RegisterWatcher(w)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestStoreManagerConcurrency(t *testing.T) {
	sm := &StoreManager{modules: make(map[string]map[string]*Store)}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(i int) {
			sm.RegisterStore("m", fmt.Sprintf("s%d", i), &Store{state: make(map[string]any)})
			wg.Done()
		}(i)
		go func(i int) {
			_ = sm.GetStore("m", fmt.Sprintf("s%d", i))
			wg.Done()
		}(i)
		go func() {
			_ = sm.Snapshot()
			wg.Done()
		}()
	}
	wg.Wait()
}
