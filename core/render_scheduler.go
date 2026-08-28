//go:build js && wasm

package core

import (
	"fmt"
	"sort"
	"sync"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/internal/rendertrace"
	"github.com/rfwlab/rfw/v2/js"
)

var componentRenderScheduler = struct {
	sync.Mutex
	pending   map[string]renderJob
	scheduled bool
	batchID   uint64
}{pending: make(map[string]renderJob)}

type renderJob struct {
	id       string
	depth    int
	active   func() bool
	evaluate func() string
	commit   func(string)
	trace    *renderJobTrace
}

type renderJobTrace struct {
	name       string
	parentID   string
	batchID    uint64
	renderID   uint64
	causes     []rendertrace.Cause
	coalesced  int
	queueDepth int
}

// requestRender invalidates a mounted component and coalesces its DOM work with
// every other request made in the same JavaScript turn. Rendering is delayed,
// not state capture: the flush always observes the latest store and signal data.
func (c *HTMLComponent) requestRender(causes ...rendertrace.Cause) {
	if c == nil || !c.mounted {
		return
	}
	c.Invalidate()
	if len(causes) == 0 {
		causes = []rendertrace.Cause{{Kind: "explicit"}}
	}
	parentID := ""
	if c.parent != nil {
		parentID = c.parent.ID
	}
	requestScheduledRender(renderJob{
		id:    c.ID,
		depth: componentDepth(c),
		active: func() bool {
			return c.mounted
		},
		evaluate: c.RenderFresh,
		commit:   func(html string) { dom.UpdateMountedDOM(c.ID, html) },
		trace:    newRenderJobTrace(c.Name, parentID, causes),
	})
}

func newRenderJobTrace(name, parentID string, causes []rendertrace.Cause) *renderJobTrace {
	if !rendertrace.Enabled() {
		return nil
	}
	return &renderJobTrace{name: name, parentID: parentID, causes: causes}
}

func requestScheduledRender(job renderJob) {
	if job.id == "" || job.active == nil || job.evaluate == nil || job.commit == nil || !job.active() {
		return
	}
	if !rendertrace.Enabled() {
		componentRenderScheduler.Lock()
		componentRenderScheduler.pending[job.id] = job
		if componentRenderScheduler.scheduled {
			componentRenderScheduler.Unlock()
			return
		}
		componentRenderScheduler.scheduled = true
		componentRenderScheduler.batchID = 0
		componentRenderScheduler.Unlock()
		scheduleComponentRenderFlush()
		return
	}
	if job.trace == nil {
		job.trace = &renderJobTrace{}
	}
	if len(job.trace.causes) == 0 {
		job.trace.causes = []rendertrace.Cause{{Kind: "explicit"}}
	}
	for index := range job.trace.causes {
		job.trace.causes[index] = rendertrace.NormalizeCause(job.trace.causes[index])
	}

	componentRenderScheduler.Lock()
	if existing, ok := componentRenderScheduler.pending[job.id]; ok {
		existing.active = job.active
		existing.evaluate = job.evaluate
		existing.commit = job.commit
		existing.depth = job.depth
		if existing.trace == nil {
			existing.trace = &renderJobTrace{}
		}
		existing.trace.name = job.trace.name
		existing.trace.parentID = job.trace.parentID
		existing.trace.coalesced++
		for _, cause := range job.trace.causes {
			existing.trace.causes = rendertrace.AppendCause(existing.trace.causes, cause)
		}
		existing.trace.queueDepth = len(componentRenderScheduler.pending)
		componentRenderScheduler.pending[job.id] = existing
		eventJob := cloneRenderJobTrace(existing)
		componentRenderScheduler.Unlock()
		emitRenderJob("coalesced", eventJob, rendertrace.NormalizeCause(job.trace.causes[0]), "", "")
		return
	}

	needsCallback := !componentRenderScheduler.scheduled
	if needsCallback {
		componentRenderScheduler.scheduled = true
		componentRenderScheduler.batchID = rendertrace.NextBatchID()
	}
	job.trace.batchID = componentRenderScheduler.batchID
	job.trace.renderID = rendertrace.NextRenderID()
	componentRenderScheduler.pending[job.id] = job
	job.trace.queueDepth = len(componentRenderScheduler.pending)
	componentRenderScheduler.pending[job.id] = job
	eventJob := cloneRenderJobTrace(job)
	componentRenderScheduler.Unlock()
	emitRenderJob("scheduled", eventJob, eventJob.trace.causes[0], "", "")
	if !needsCallback {
		return
	}
	scheduleComponentRenderFlush()
}

func cloneRenderJobTrace(job renderJob) renderJob {
	if job.trace == nil {
		return job
	}
	trace := *job.trace
	trace.causes = append([]rendertrace.Cause(nil), trace.causes...)
	job.trace = &trace
	return job
}

func scheduleComponentRenderFlush() {
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
	job, pending := componentRenderScheduler.pending[componentID]
	if pending {
		delete(componentRenderScheduler.pending, componentID)
	}
	componentRenderScheduler.Unlock()
	if pending {
		if !rendertrace.Enabled() {
			return
		}
		emitRenderJob("cancelled", job, firstCause(job.trace.causes), "cancelled", "unmounted")
	}
}

func flushComponentRenders() {
	componentRenderScheduler.Lock()
	pending := make([]renderJob, 0, len(componentRenderScheduler.pending))
	for _, job := range componentRenderScheduler.pending {
		pending = append(pending, job)
	}
	clear(componentRenderScheduler.pending)
	componentRenderScheduler.scheduled = false
	componentRenderScheduler.batchID = 0
	componentRenderScheduler.Unlock()

	// Parents commit before descendants. Ownership boundaries keep a parent from
	// touching child internals; the ordering makes both renders deterministic.
	sort.Slice(pending, func(i, j int) bool {
		if pending[i].depth != pending[j].depth {
			return pending[i].depth < pending[j].depth
		}
		return pending[i].id < pending[j].id
	})
	for _, job := range pending {
		if !job.active() {
			if rendertrace.Enabled() {
				emitRenderJob("cancelled", job, firstCause(job.trace.causes), "cancelled", "inactive")
			}
			continue
		}
		runRenderJob(job)
	}
}

func runRenderJob(job renderJob) {
	if !rendertrace.Enabled() {
		job.commit(job.evaluate())
		return
	}
	trace := job.trace
	started := rendertrace.NowMS()
	emitRenderJob("started", job, firstCause(trace.causes), "", "")
	var templateMS, domMS float64
	phase, phaseStarted := "template", rendertrace.NowMS()
	defer func() {
		if recovered := recover(); recovered != nil {
			now := rendertrace.NowMS()
			switch phase {
			case "template":
				templateMS = now - phaseStarted
			case "dom":
				domMS = now - phaseStarted
			}
			rendertrace.Emit(rendertrace.Record{
				Event:             "failed",
				BatchID:           trace.batchID,
				RenderID:          trace.renderID,
				ComponentID:       job.id,
				ComponentName:     trace.name,
				ParentComponentID: trace.parentID,
				Depth:             job.depth,
				Cause:             firstCause(trace.causes),
				Causes:            trace.causes,
				QueueDepth:        trace.queueDepth,
				CoalescedCount:    trace.coalesced,
				TemplateMS:        templateMS,
				DOMMS:             domMS,
				TotalMS:           now - started,
				Outcome:           "failed",
				Reason:            fmt.Sprint(recovered),
			})
			panic(recovered)
		}
	}()

	html := job.evaluate()
	templateMS = rendertrace.NowMS() - phaseStarted
	phase, phaseStarted = "dom", rendertrace.NowMS()
	job.commit(html)
	domMS = rendertrace.NowMS() - phaseStarted
	phase = ""
	rendertrace.Emit(rendertrace.Record{
		Event:             "committed",
		BatchID:           trace.batchID,
		RenderID:          trace.renderID,
		ComponentID:       job.id,
		ComponentName:     trace.name,
		ParentComponentID: trace.parentID,
		Depth:             job.depth,
		Cause:             firstCause(trace.causes),
		Causes:            trace.causes,
		QueueDepth:        trace.queueDepth,
		CoalescedCount:    trace.coalesced,
		TemplateMS:        templateMS,
		DOMMS:             domMS,
		TotalMS:           rendertrace.NowMS() - started,
		Outcome:           "committed",
	})
}

func emitRenderJob(event string, job renderJob, cause rendertrace.Cause, outcome, reason string) {
	trace := job.trace
	rendertrace.Emit(rendertrace.Record{
		Event:             event,
		BatchID:           trace.batchID,
		RenderID:          trace.renderID,
		ComponentID:       job.id,
		ComponentName:     trace.name,
		ParentComponentID: trace.parentID,
		Depth:             job.depth,
		Cause:             cause,
		Causes:            trace.causes,
		QueueDepth:        trace.queueDepth,
		CoalescedCount:    trace.coalesced,
		Outcome:           outcome,
		Reason:            reason,
	})
}

func firstCause(causes []rendertrace.Cause) rendertrace.Cause {
	if len(causes) == 0 {
		return rendertrace.Cause{Kind: "explicit"}
	}
	return rendertrace.NormalizeCause(causes[0])
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
