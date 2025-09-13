//go:build js && wasm

package toast

import (
	"encoding/json"
	"sync"
	"time"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
)

// Plugin displays temporary toast notifications stacked in the bottom right.
type Plugin struct {
	DefaultDuration time.Duration
	// Template renders the toast element. When nil, a default template is used.
	Template  func(string, []Action, func()) dom.Element
	queue     chan item
	container dom.Element
}

type item struct {
	msg      string
	dur      time.Duration
	actions  []Action
	template func(string, []Action, func()) dom.Element
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
		once := sync.Once{}
		var el dom.Element
		close := func() { once.Do(func() { el.Call("remove") }) }

		tpl := it.template
		if tpl == nil {
			tpl = p.Template
		}
		if tpl == nil {
			tpl = defaultTemplate
		}
		el = tpl(it.msg, it.actions, close)

		p.container.Call("appendChild", el.Value)
		d := it.dur
		if d == 0 {
			d = p.DefaultDuration
		}
		time.AfterFunc(d, close)
	}
}

var plug *Plugin

// Push enqueues a message using the default duration.
func Push(msg string) { PushTimed(msg, 0) }

// PushTimed enqueues a message with a custom duration.
func PushTimed(msg string, d time.Duration) {
	PushOptions(msg, Options{Duration: d})
}

// Action represents an optional button displayed with the toast.
type Action struct {
	Label   string
	Handler func()
}

// Options configure a toast message.
type Options struct {
	Duration time.Duration
	Actions  []Action
	Template func(string, []Action, func()) dom.Element
}

// PushOptions enqueues a message with additional configuration.
func PushOptions(msg string, opts Options) {
	if plug == nil {
		return
	}
	plug.queue <- item{msg: msg, dur: opts.Duration, actions: opts.Actions, template: opts.Template}
}

func defaultTemplate(msg string, actions []Action, close func()) dom.Element {
	doc := dom.Doc()
	el := doc.CreateElement("div")
	el.AddClass("bg-gray-800")
	el.AddClass("text-white")
	el.AddClass("px-4")
	el.AddClass("py-2")
	el.AddClass("mb-2")
	el.AddClass("rounded")

	body := doc.CreateElement("div")
	body.SetText(msg)
	el.Call("appendChild", body.Value)

	btnWrap := doc.CreateElement("div")
	btnWrap.AddClass("flex")
	btnWrap.AddClass("mt-2")

	for _, a := range actions {
		btn := doc.CreateElement("button")
		btn.SetText(a.Label)
		btn.AddClass("bg-blue-500")
		btn.AddClass("text-white")
		btn.AddClass("px-2")
		btn.AddClass("py-1")
		btn.AddClass("rounded")
		btn.AddClass("mr-2")
		btn.OnClick(func(dom.Event) { a.Handler(); close() })
		btnWrap.Call("appendChild", btn.Value)
	}

	closeBtn := doc.CreateElement("button")
	closeBtn.SetText("×")
	closeBtn.AddClass("ml-auto")
	closeBtn.OnClick(func(dom.Event) { close() })

	btnWrap.Call("appendChild", closeBtn.Value)
	el.Call("appendChild", btnWrap.Value)
	return el
}
