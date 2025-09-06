package core

import (
	"fmt"
	"sync"
	"testing"
)

func TestComponentRegistryConcurrentAccess(t *testing.T) {
	componentRegistryMu.Lock()
	ComponentRegistry = map[string]func() Component{}
	componentRegistryMu.Unlock()

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("comp-%d", i)
			if err := RegisterComponent(name, func() Component { return noopComponent{} }); err != nil {
				t.Errorf("register %s: %v", name, err)
			}
			if c := LoadComponent(name); c == nil {
				t.Errorf("load %s: got nil", name)
			}
		}(i)
	}
	wg.Wait()
}
