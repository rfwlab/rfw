package devtools

import (
	"sort"
	"sync"
	"time"
)

type lifecycleEvent struct {
	Kind string
	At   time.Time
}

const lifecycleLimit = 64

var (
	lifecycleMu          sync.RWMutex
	lifecycleByComponent = map[string][]lifecycleEvent{}
)

func appendLifecycle(id, kind string, at time.Time) {
	if id == "" || kind == "" {
		return
	}
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()
	entry := lifecycleEvent{Kind: kind, At: at}
	list := append(lifecycleByComponent[id], entry)
	if len(list) > lifecycleLimit {
		list = append([]lifecycleEvent(nil), list[len(list)-lifecycleLimit:]...)
	}
	lifecycleByComponent[id] = list
}

func dropLifecycle(id string) {
	if id == "" {
		return
	}
	lifecycleMu.Lock()
	delete(lifecycleByComponent, id)
	lifecycleMu.Unlock()
}

func snapshotLifecycle(id string) []lifecycleEvent {
	lifecycleMu.RLock()
	list := lifecycleByComponent[id]
	lifecycleMu.RUnlock()
	if len(list) == 0 {
		return nil
	}
	out := make([]lifecycleEvent, len(list))
	copy(out, list)
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}

func resetLifecycles() {
	lifecycleMu.Lock()
	lifecycleByComponent = map[string][]lifecycleEvent{}
	lifecycleMu.Unlock()
}
