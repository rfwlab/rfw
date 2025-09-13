//go:build js && wasm

package toast

import (
	"encoding/json"
	"time"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
)

// Plugin displays temporary toast notifications stacked in the bottom right.
type Plugin struct {
	DefaultDuration time.Duration
	queue           chan item
	container       dom.Element
}

type item struct {
	msg string
	dur time.Duration
}

// New constructs a toast plugin with a default display duration.
func New() *Plugin {
	return &Plugin{DefaultDuration: 3 * time.Second, queue: make(chan item, 8)}
}

func (p *Plugin) Build(json.RawMessage) error { return nil }

// Install prepares the toast container and starts processing the queue.
func (p *Plugin) Install(a *core.App) {
	plug = p
	doc := dom.Doc()
	body := doc.Query("body")
	container := doc.CreateElement("div")
	container.SetStyle("position", "fixed")
	container.SetStyle("bottom", "1rem")
	container.SetStyle("right", "1rem")
	container.SetStyle("display", "flex")
	container.SetStyle("flex-direction", "column")
	body.Call("appendChild", container.Value)
	p.container = container
	go p.loop()
}

func (p *Plugin) loop() {
	for it := range p.queue {
		el := dom.Doc().CreateElement("div")
		el.SetText(it.msg)
		el.AddClass("bg-gray-800")
		el.AddClass("text-white")
		el.AddClass("px-4")
		el.AddClass("py-2")
		el.AddClass("mb-2")
		el.AddClass("rounded")
		p.container.Call("appendChild", el.Value)
		d := it.dur
		if d == 0 {
			d = p.DefaultDuration
		}
		time.Sleep(d)
		el.Call("remove")
	}
}

var plug *Plugin

// Push enqueues a message using the default duration.
func Push(msg string) { PushTimed(msg, 0) }

// PushTimed enqueues a message with a custom duration.
func PushTimed(msg string, d time.Duration) {
	if plug == nil {
		return
	}
	plug.queue <- item{msg: msg, dur: d}
}
