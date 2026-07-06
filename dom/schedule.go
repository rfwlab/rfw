//go:build js && wasm

package dom

import (
	"sync"
	"time"
)

var sched = struct {
	sync.Mutex
	timers map[string]*time.Timer
}{timers: make(map[string]*time.Timer)}

// ScheduleRender updates the DOM of the specified component after a delay.
func ScheduleRender(componentID string, html string, delay time.Duration) {
	sched.Lock()
	defer sched.Unlock()
	if t, ok := sched.timers[componentID]; ok {
		t.Stop()
	}
	sched.timers[componentID] = time.AfterFunc(delay, func() {
		UpdateDOM(componentID, html)
		sched.Lock()
		delete(sched.timers, componentID)
		sched.Unlock()
	})
}
