//go:build js && wasm

package dom

import (
	"sync"
	"sync/atomic"
)

// LifecycleHook observes a component root after mount, update, and unmount.
type LifecycleHook struct {
	Mounted   func(Element) func()
	Updated   func(Element)
	Unmounted func(Element)
}

type lifecycleRecord struct {
	id      uint64
	hook    LifecycleHook
	mu      sync.Mutex
	cleanup func()
	stopped bool
}

func (record *lifecycleRecord) setCleanup(cleanup func()) {
	record.mu.Lock()
	if record.stopped {
		record.mu.Unlock()
		if cleanup != nil {
			cleanup()
		}
		return
	}
	record.cleanup = cleanup
	record.mu.Unlock()
}

func (record *lifecycleRecord) takeCleanup(stop bool) func() {
	record.mu.Lock()
	cleanup := record.cleanup
	record.cleanup = nil
	if stop {
		record.stopped = true
	}
	record.mu.Unlock()
	return cleanup
}

type componentLifecycle struct {
	mounted bool
	hooks   []*lifecycleRecord
}

var lifecycleHooks = struct {
	sync.Mutex
	sequence   atomic.Uint64
	components map[string]*componentLifecycle
}{components: make(map[string]*componentLifecycle)}

// RegisterLifecycleHook registers a hook and returns a cancellation function.
func RegisterLifecycleHook(componentID string, hook LifecycleHook) func() {
	record := &lifecycleRecord{id: lifecycleHooks.sequence.Add(1), hook: hook}
	lifecycleHooks.Lock()
	component := lifecycleHooks.components[componentID]
	if component == nil {
		component = &componentLifecycle{}
		lifecycleHooks.components[componentID] = component
	}
	component.hooks = append(component.hooks, record)
	mounted := component.mounted
	lifecycleHooks.Unlock()

	if mounted && hook.Mounted != nil {
		record.setCleanup(runMountedHook(componentID, hook.Mounted, ownComponentRoot(componentID)))
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			lifecycleHooks.Lock()
			component := lifecycleHooks.components[componentID]
			if component != nil {
				for index, candidate := range component.hooks {
					if candidate.id == record.id {
						component.hooks = append(component.hooks[:index], component.hooks[index+1:]...)
						break
					}
				}
				if len(component.hooks) == 0 {
					delete(lifecycleHooks.components, componentID)
				}
			}
			cleanup := record.takeCleanup(true)
			lifecycleHooks.Unlock()
			if cleanup != nil {
				cleanup()
			}
		})
	}
}

// MountLifecycleHooks activates every hook registered for a component.
func MountLifecycleHooks(componentID string) {
	lifecycleHooks.Lock()
	component := lifecycleHooks.components[componentID]
	if component == nil || component.mounted {
		lifecycleHooks.Unlock()
		return
	}
	component.mounted = true
	hooks := append([]*lifecycleRecord(nil), component.hooks...)
	lifecycleHooks.Unlock()
	root := ownComponentRoot(componentID)
	for _, record := range hooks {
		if record.hook.Mounted != nil {
			record.setCleanup(runMountedHook(componentID, record.hook.Mounted, root))
		}
	}
}

// UpdateLifecycleHooks notifies mounted component hooks after a DOM patch.
func UpdateLifecycleHooks(componentID string) {
	lifecycleHooks.Lock()
	component := lifecycleHooks.components[componentID]
	if component == nil || !component.mounted {
		lifecycleHooks.Unlock()
		return
	}
	hooks := append([]*lifecycleRecord(nil), component.hooks...)
	lifecycleHooks.Unlock()
	root := ownComponentRoot(componentID)
	for _, record := range hooks {
		if record.hook.Updated != nil {
			runLifecycleHook(componentID, "updated", func() {
				record.hook.Updated(root)
			})
		}
	}
}

// UnmountLifecycleHooks runs hook cleanup while the component root still exists.
func UnmountLifecycleHooks(componentID string) {
	lifecycleHooks.Lock()
	component := lifecycleHooks.components[componentID]
	if component == nil || !component.mounted {
		lifecycleHooks.Unlock()
		return
	}
	component.mounted = false
	hooks := append([]*lifecycleRecord(nil), component.hooks...)
	lifecycleHooks.Unlock()
	root := ownComponentRoot(componentID)
	for index := len(hooks) - 1; index >= 0; index-- {
		record := hooks[index]
		if cleanup := record.takeCleanup(false); cleanup != nil {
			runLifecycleHook(componentID, "cleanup", cleanup)
		}
		if record.hook.Unmounted != nil {
			runLifecycleHook(componentID, "unmounted", func() {
				record.hook.Unmounted(root)
			})
		}
	}
}

func runMountedHook(componentID string, hook func(Element) func(), root Element) (cleanup func()) {
	runLifecycleHook(componentID, "mounted", func() {
		cleanup = hook(root)
	})
	return cleanup
}

func runLifecycleHook(componentID, phase string, fn func()) {
	defer func() {
		if recovered := recover(); recovered != nil && OnHandlerPanic != nil {
			OnHandlerPanic(recovered, "DOM "+phase+": "+componentID)
		}
	}()
	fn()
}

func ownComponentRoot(componentID string) Element {
	root := ComponentRoot(componentID)
	if root.IsNull() || root.IsUndefined() || root.Attr("data-component-id") != componentID {
		return Element{}
	}
	return root
}
