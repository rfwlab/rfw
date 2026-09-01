//go:build js && wasm

// Package devtools provides an in-page developer tools overlay for rfw
// applications. Once installed, pressing Ctrl+Shift+D toggles a fixed
// panel with two tabs:
//
//   - Components: a tree of live components, walked from the router's
//     current component through HTMLComponent.Dependencies, plus any
//     other components observed through App lifecycle hooks.
//   - Stores: all stores registered on the global state.StoreManager,
//     with their key/values JSON-stringified (truncated to 200 chars).
//
// Known limitations:
//
//   - core's internal dev component registry (devRegisterComponent) is
//     unexported and compiled as a no-op in js/wasm builds, so there is no
//     framework-level enumeration of every component ever created. The
//     plugin instead tracks live components via core.App.RegisterLifecycle
//     (mount/unmount hooks) and walks the Dependencies tree from
//     router.CurrentComponent(). Components mounted before the plugin was
//     installed and not reachable from the current route are not listed.
//   - Store enumeration relies on state.GlobalStoreManager.Snapshot(),
//     which copies module/store/key state. Signals and computed values that
//     are not backed by a store key are not visible.
//
// The plugin has zero cost when not installed: it only registers hooks and
// builds its overlay DOM lazily on first toggle.
package devtools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/plugins/shortcut"
	"github.com/rfwlab/rfw/v2/router"
	"github.com/rfwlab/rfw/v2/state"
)

// valueLimit is the maximum number of characters shown for a single
// JSON-stringified store value.
const valueLimit = 200

const (
	tabComponents = "components"
	tabStores     = "stores"
)

// Plugin renders the devtools overlay. Create it with New and register it
// through core.RegisterPlugin.
type Plugin struct {
	// Shortcut is the key combination toggling the overlay.
	// Defaults to "control+shift+d".
	Shortcut string
	// Manager is the store manager inspected by the Stores tab.
	// Defaults to state.GlobalStoreManager.
	Manager *state.StoreManager

	mounted map[string]core.Component
	panel   dom.Element
	content dom.Element
	tabBtns map[string]dom.Element
	tab     string
	built   bool
	visible bool
}

// New creates a devtools plugin with default settings.
func New() *Plugin {
	return &Plugin{
		Shortcut: "control+shift+d",
		Manager:  state.GlobalStoreManager,
		mounted:  make(map[string]core.Component),
		tab:      tabComponents,
	}
}

// Name identifies the plugin for deduplication.
func (p *Plugin) Name() string { return "devtools" }

// Build is a no-op; devtools has no build step.
func (p *Plugin) Build(json.RawMessage) error { return nil }

// Optional declares the shortcut plugin as an optional dependency so the
// toggle keybinding works out of the box.
func (p *Plugin) Optional() []core.Plugin { return []core.Plugin{shortcut.New()} }

// Install registers lifecycle hooks to track live components and binds the
// toggle shortcut. The overlay DOM is created lazily on first toggle.
func (p *Plugin) Install(a *core.App) {
	if p.Manager == nil {
		p.Manager = state.GlobalStoreManager
	}
	if p.mounted == nil {
		p.mounted = make(map[string]core.Component)
	}
	if p.tab == "" {
		p.tab = tabComponents
	}
	a.RegisterLifecycle(
		func(c core.Component) {
			p.mounted[c.GetID()] = c
			if p.visible {
				p.refresh()
			}
		},
		func(c core.Component) {
			delete(p.mounted, c.GetID())
			if p.visible {
				p.refresh()
			}
		},
	)
	if p.Shortcut != "" {
		shortcut.Bind(p.Shortcut, p.Toggle)
	}
}

// Toggle shows or hides the overlay panel.
func (p *Plugin) Toggle() {
	if p.visible {
		p.Hide()
	} else {
		p.Show()
	}
}

// Show opens the overlay panel and refreshes the active tab.
func (p *Plugin) Show() {
	p.ensurePanel()
	p.panel.SetStyle("display", "flex")
	p.visible = true
	p.refresh()
}

// Hide closes the overlay panel.
func (p *Plugin) Hide() {
	if !p.built {
		return
	}
	p.panel.SetStyle("display", "none")
	p.visible = false
}

// ensurePanel builds the overlay DOM once and appends it to the body.
func (p *Plugin) ensurePanel() {
	if p.built {
		return
	}
	doc := dom.Doc()
	panel := doc.CreateElement("div")
	panel.SetAttr("id", "rfw-devtools")
	for prop, val := range map[string]string{
		"position":       "fixed",
		"top":            "0",
		"right":          "0",
		"width":          "380px",
		"max-height":     "70vh",
		"display":        "flex",
		"flex-direction": "column",
		"background":     "#101010",
		"color":          "#d4d4d4",
		"font-family":    "monospace",
		"font-size":      "12px",
		"border":         "1px solid #333",
		"border-radius":  "0 0 0 6px",
		"z-index":        "2147483000",
		"box-shadow":     "0 2px 12px rgba(0,0,0,.6)",
	} {
		panel.SetStyle(prop, val)
	}

	header := doc.CreateElement("div")
	header.SetStyle("display", "flex")
	header.SetStyle("align-items", "center")
	header.SetStyle("gap", "6px")
	header.SetStyle("padding", "6px 8px")
	header.SetStyle("border-bottom", "1px solid #333")

	title := doc.CreateElement("span")
	title.SetText("rfw devtools")
	title.SetStyle("font-weight", "bold")
	title.SetStyle("margin-right", "auto")
	header.AppendChild(title)

	p.tabBtns = make(map[string]dom.Element)
	header.AppendChild(p.tabButton(doc, "Components", tabComponents))
	header.AppendChild(p.tabButton(doc, "Stores", tabStores))

	refresh := headerButton(doc, "Refresh")
	refresh.OnClick(func(dom.Event) { p.refresh() })
	header.AppendChild(refresh)

	closeBtn := headerButton(doc, "X")
	closeBtn.OnClick(func(dom.Event) { p.Hide() })
	header.AppendChild(closeBtn)

	content := doc.CreateElement("div")
	content.SetAttr("id", "rfw-devtools-content")
	content.SetStyle("overflow", "auto")
	content.SetStyle("padding", "8px")

	panel.AppendChild(header)
	panel.AppendChild(content)
	doc.Query("body").AppendChild(panel)

	p.panel = panel
	p.content = content
	p.built = true
	p.highlightTabs()
}

func (p *Plugin) tabButton(doc dom.Document, label, tab string) dom.Element {
	btn := headerButton(doc, label)
	btn.OnClick(func(dom.Event) {
		p.tab = tab
		p.highlightTabs()
		p.refresh()
	})
	p.tabBtns[tab] = btn
	return btn
}

func headerButton(doc dom.Document, label string) dom.Element {
	btn := doc.CreateElement("button")
	btn.SetText(label)
	btn.SetStyle("background", "#1c1c1c")
	btn.SetStyle("color", "#d4d4d4")
	btn.SetStyle("border", "1px solid #444")
	btn.SetStyle("border-radius", "3px")
	btn.SetStyle("font", "inherit")
	btn.SetStyle("padding", "2px 6px")
	btn.SetStyle("cursor", "pointer")
	return btn
}

func (p *Plugin) highlightTabs() {
	for tab, btn := range p.tabBtns {
		if tab == p.tab {
			btn.SetStyle("border-color", "#888")
			btn.SetStyle("background", "#2a2a2a")
		} else {
			btn.SetStyle("border-color", "#444")
			btn.SetStyle("background", "#1c1c1c")
		}
	}
}

// refresh re-renders the active tab into the panel content area.
func (p *Plugin) refresh() {
	if !p.built {
		return
	}
	p.content.SetHTML("")
	switch p.tab {
	case tabStores:
		p.renderStores(p.content)
	default:
		p.renderComponents(p.content)
	}
}

// renderComponents renders the component tree, rooted at the router's
// current component, followed by any other tracked mounted components.
func (p *Plugin) renderComponents(into dom.Element) {
	seen := make(map[string]bool)
	var roots []core.Component
	if c := router.CurrentComponent(); c != nil {
		roots = append(roots, c)
	}
	if len(roots) > 0 {
		buildComponentTree(into, roots, seen)
	}

	var extra []core.Component
	for id, c := range p.mounted {
		if !seen[id] {
			extra = append(extra, c)
		}
	}
	if len(extra) > 0 {
		sort.Slice(extra, func(i, j int) bool { return extra[i].GetID() < extra[j].GetID() })
		label := dom.Doc().CreateElement("div")
		label.SetText("Other mounted components:")
		label.SetStyle("margin-top", "6px")
		label.SetStyle("color", "#888")
		into.AppendChild(label)
		buildComponentTree(into, extra, seen)
	}

	if len(roots) == 0 && len(extra) == 0 {
		empty := dom.Doc().CreateElement("div")
		empty.SetText("No components found (no routed component and no mount events observed).")
		empty.SetStyle("color", "#888")
		into.AppendChild(empty)
	}
}

// buildComponentTree appends a nested list describing roots and their
// Dependencies subtrees to into. Visited component IDs are recorded in seen
// to guard against cycles and duplicates.
func buildComponentTree(into dom.Element, roots []core.Component, seen map[string]bool) {
	doc := dom.Doc()
	ul := doc.CreateElement("ul")
	ul.SetStyle("margin", "0")
	ul.SetStyle("padding-left", "16px")
	ul.SetStyle("list-style", "square")
	for _, c := range roots {
		appendComponentNode(doc, ul, c, seen)
	}
	into.AppendChild(ul)
}

func appendComponentNode(doc dom.Document, parent dom.Element, c core.Component, seen map[string]bool) {
	li := doc.CreateElement("li")
	li.SetStyle("margin", "2px 0")
	label := doc.CreateElement("span")
	label.SetText(fmt.Sprintf("%s (%s)", c.GetName(), c.GetID()))
	li.AppendChild(label)
	parent.AppendChild(li)

	id := c.GetID()
	if seen[id] {
		return
	}
	seen[id] = true

	hc := unwrapHTML(c)
	if hc == nil || len(hc.Dependencies) == 0 {
		return
	}
	names := make([]string, 0, len(hc.Dependencies))
	for name := range hc.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)

	ul := doc.CreateElement("ul")
	ul.SetStyle("margin", "0")
	ul.SetStyle("padding-left", "16px")
	ul.SetStyle("list-style", "square")
	for _, name := range names {
		appendComponentNode(doc, ul, hc.Dependencies[name], seen)
	}
	li.AppendChild(ul)
}

// unwrapHTML extracts the underlying *core.HTMLComponent, if any, to access
// the Dependencies tree. Composition components expose Unwrap.
func unwrapHTML(c core.Component) *core.HTMLComponent {
	switch v := c.(type) {
	case *core.HTMLComponent:
		return v
	case interface{ Unwrap() *core.HTMLComponent }:
		return v.Unwrap()
	}
	return nil
}

// renderStores renders every module/store/key of the configured store
// manager, with values JSON-stringified and truncated.
func (p *Plugin) renderStores(into dom.Element) {
	doc := dom.Doc()
	snap := p.Manager.Snapshot()
	if len(snap) == 0 {
		empty := doc.CreateElement("div")
		empty.SetText("No stores registered.")
		empty.SetStyle("color", "#888")
		into.AppendChild(empty)
		return
	}

	modules := make([]string, 0, len(snap))
	for m := range snap {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	for _, module := range modules {
		stores := snap[module]
		names := make([]string, 0, len(stores))
		for n := range stores {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, name := range names {
			heading := doc.CreateElement("div")
			heading.SetText(module + "/" + name)
			heading.SetStyle("font-weight", "bold")
			heading.SetStyle("margin", "6px 0 2px")
			heading.SetStyle("border-bottom", "1px solid #333")
			into.AppendChild(heading)

			kv := stores[name]
			keys := make([]string, 0, len(kv))
			for k := range kv {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			if len(keys) == 0 {
				row := doc.CreateElement("div")
				row.SetText("(empty)")
				row.SetStyle("color", "#888")
				into.AppendChild(row)
				continue
			}
			for _, k := range keys {
				row := doc.CreateElement("div")
				row.SetStyle("white-space", "pre-wrap")
				row.SetStyle("word-break", "break-all")
				row.SetText(k + " = " + stringify(kv[k]))
				into.AppendChild(row)
			}
		}
	}
}

// stringify JSON-encodes v, falling back to fmt formatting, and truncates
// the result to valueLimit characters.
func stringify(v any) string {
	s := ""
	if b, err := json.Marshal(v); err == nil {
		s = string(b)
	} else {
		s = fmt.Sprintf("%v", v)
	}
	if len(s) > valueLimit {
		s = s[:valueLimit] + "…"
	}
	return s
}
