//go:build js && wasm

package core

import (
	"sort"
	"sync"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/js"
)

var componentRenderScheduler = struct {
	sync.Mutex
	pending   map[string]renderJob
	scheduled bool
}{pending: make(map[string]renderJob)}

type renderJob struct {
	id     string
	depth  int
	active func() bool
	render func()
}

// requestRender invalidates a mounted component and coalesces its DOM work with
// every other request made in the same JavaScript turn. Rendering is delayed,
// not state capture: the flush always observes the latest store and signal data.
func (c *HTMLComponent) requestRender() {
	if c == nil || !c.mounted {
		return
	}
	c.Invalidate()
	requestScheduledRender(renderJob{
		id:    c.ID,
		depth: componentDepth(c),
		active: func() bool {
			return c.mounted
		},
		render: func() {
			dom.UpdateMountedDOM(c.ID, c.RenderFresh())
		},
	})
}

func requestScheduledRender(job renderJob) {
	if job.id == "" || job.active == nil || job.render == nil || !job.active() {
		return
	}
	componentRenderScheduler.Lock()
	componentRenderScheduler.pending[job.id] = job
	if componentRenderScheduler.scheduled {
		componentRenderScheduler.Unlock()
		return
	}
	componentRenderScheduler.scheduled = true
	componentRenderScheduler.Unlock()

	var callback js.Func
	callback = js.SafeFuncOf(func(js.Value, []js.Value) any {
		defer callback.Release()
		flushComponentRenders()
		return nil
	})
	queueMicrotask := js.Global().Get("queueMicrotask")
	if queueMicrotask.Type() == js.TypeFunction {
		queueMicrotask.Invoke(callback.Value)
		return
	}
	js.Global().Call("setTimeout", callback.Value, 0)
}

func cancelComponentRender(componentID string) {
	componentRenderScheduler.Lock()
	delete(componentRenderScheduler.pending, componentID)
	componentRenderScheduler.Unlock()
}

func flushComponentRenders() {
	componentRenderScheduler.Lock()
	pending := make([]renderJob, 0, len(componentRenderScheduler.pending))
	for _, job := range componentRenderScheduler.pending {
		pending = append(pending, job)
	}
	clear(componentRenderScheduler.pending)
	componentRenderScheduler.scheduled = false
	componentRenderScheduler.Unlock()

	// Parents commit before descendants. Ownership boundaries keep a parent from
	// touching child internals; the ordering makes both renders deterministic.
	sort.SliceStable(pending, func(i, j int) bool {
		return pending[i].depth < pending[j].depth
	})
	for _, job := range pending {
		if !job.active() {
			continue
		}
		job.render()
	}
}

func componentDepth(component *HTMLComponent) int {
	depth := 0
	for component != nil && component.parent != nil {
		depth++
		component = component.parent
	}
	return depth
}

func mountedComponentDepth(componentID string) int {
	root := dom.ComponentRoot(componentID)
	if root.IsNull() || root.IsUndefined() {
		return 0
	}
	depth := 0
	for parent := root.Get("parentElement"); parent.Truthy(); parent = parent.Get("parentElement") {
		if parent.Call("hasAttribute", "data-component-id").Bool() {
			depth++
		}
	}
	return depth
}
