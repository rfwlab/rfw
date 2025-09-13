//go:build js && wasm

package seo

import (
	"encoding/json"
	"fmt"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
)

type Plugin struct{}

func New() *Plugin { return &Plugin{} }

var (
	titleEl    dom.Element
	metaEls    map[string]dom.Element
	cfgTitle   string
	cfgMeta    map[string]string
	cfgPattern string
)

// SetTitle updates the document title. Override before RegisterPlugin for custom behavior.
var SetTitle func(string) = defaultSetTitle

// SetMeta updates or creates a meta tag. Override before RegisterPlugin for custom behavior.
var SetMeta func(string, string) = defaultSetMeta

func (p *Plugin) Name() string { return "seo" }

func (p *Plugin) Install(a *core.App) {
	doc := dom.Doc()
	head := doc.Head()
	titleEl = head.Query("title")
	if !titleEl.Truthy() {
		titleEl = doc.CreateElement("title")
		head.AppendChild(titleEl)
	}
	if cfgTitle != "" {
		titleEl.SetText(cfgTitle)
	}
	metaEls = make(map[string]dom.Element)
	for name, content := range cfgMeta {
		sel := fmt.Sprintf(`meta[name="%s"]`, name)
		el := head.Query(sel)
		if !el.Truthy() {
			el = doc.CreateElement("meta")
			el.SetAttr("name", name)
			head.AppendChild(el)
		}
		el.SetAttr("content", content)
		metaEls[name] = el
	}
}

func (p *Plugin) Build(cfg json.RawMessage) error {
	if len(cfg) == 0 {
		return nil
	}
	data := struct {
		Title   string            `json:"title"`
		Meta    map[string]string `json:"meta"`
		Pattern string            `json:"pattern"`
	}{}
	if err := json.Unmarshal(cfg, &data); err != nil {
		return err
	}
	cfgTitle = data.Title
	cfgMeta = data.Meta
	cfgPattern = data.Pattern
	return nil
}

func defaultSetTitle(t string) {
	if !titleEl.Truthy() {
		doc := dom.Doc()
		head := doc.Head()
		titleEl = head.Query("title")
		if !titleEl.Truthy() {
			titleEl = doc.CreateElement("title")
			head.AppendChild(titleEl)
		}
	}
	if cfgPattern != "" {
		titleEl.SetText(fmt.Sprintf(cfgPattern, t))
	} else {
		titleEl.SetText(t)
	}
}

func defaultSetMeta(name, content string) {
	if metaEls == nil {
		metaEls = make(map[string]dom.Element)
	}
	el, ok := metaEls[name]
	if !ok || !el.Truthy() {
		doc := dom.Doc()
		head := doc.Head()
		sel := fmt.Sprintf(`meta[name="%s"]`, name)
		el = head.Query(sel)
		if !el.Truthy() {
			el = doc.CreateElement("meta")
			el.SetAttr("name", name)
			head.AppendChild(el)
		}
		metaEls[name] = el
	}
	el.SetAttr("content", content)
}
