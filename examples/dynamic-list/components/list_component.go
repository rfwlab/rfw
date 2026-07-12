//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"
	"html"
	"strconv"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
)

//go:embed templates/list_component.rtml
var listTpl []byte

// items is the source of truth; the tbody is re-rendered from it. In a real
// app this slice would come from an API response (see the dynamic lists
// guide for the http.Request version).
var items = []string{"Write the docs", "Ship the release"}

type ListComponent struct {
	*core.HTMLComponent
}

func NewListComponent() *ListComponent {
	c := &ListComponent{
		HTMLComponent: core.NewHTMLComponent("ListComponent", listTpl, nil),
	}
	c.SetComponent(c)

	dom.RegisterHandlerFunc("addItem", addItem)

	// RegisterHandlerElem receives the element carrying the data-on-*
	// attribute, resolved by event delegation, so it works for rows
	// injected at runtime and one handler serves every row.
	dom.RegisterHandlerElem("removeItem", func(el dom.Element, _ dom.Event) {
		idx, err := strconv.Atoi(el.Data("idx"))
		if err != nil || idx < 0 || idx >= len(items) {
			return
		}
		items = append(items[:idx], items[idx+1:]...)
		renderItems()
	})

	c.SetOnMount(func(*core.HTMLComponent) { renderItems() })

	c.Init(nil)
	return c
}

func addItem() {
	input := dom.Query("#item-name")
	name := input.Val()
	if name == "" {
		return
	}
	items = append(items, name)
	input.SetValue("")
	renderItems()
}

// renderItems rebuilds the rows and injects them. Markup built at runtime
// uses the same @on: syntax as .rtml templates by passing it through
// dom.ExpandEvents; event delegation keeps the rows live without
// re-binding any listener.
func renderItems() {
	rows := ""
	for i, name := range items {
		rows += fmt.Sprintf(
			`<tr><td>%s</td><td><button @on:click:removeItem data-idx="%d">Remove</button></td></tr>`,
			html.EscapeString(name), i)
	}
	dom.Query("#items-rows").SetHTML(dom.ExpandEvents(rows))
}
