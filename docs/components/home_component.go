//go:build js && wasm

package components

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/rfwlab/rfw/v1/composition"
	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
	highlight "github.com/rfwlab/rfw/v1/plugins/highlight"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/home_component.rtml
var homeTpl []byte

func NewHomeComponent() *core.HTMLComponent {
	count := state.NewSignal(0)
	c := core.NewComponent("HomeComponent", homeTpl, map[string]any{"count": count})
	comp := composition.Wrap(c)

	var sparkEl dom.Element
	var cartEl dom.Element

	cart := state.NewStore("cart", state.WithHistory(20))
	if cart.Get("items") == nil {
		cart.Set("items", []string{})
	}
	state.Map(cart, "count", "items", func(v []string) int { return len(v) })

	renderDx := func() {
		dx := dom.Doc().ByID("dxGrid")
		if !dx.Truthy() {
			return
		}
		for i := 0; i < 48; i++ {
			classes := []string{"h-3", "outlined", "bg-zinc-200/60", "dark:bg-zinc-800"}
			if i%7 == 0 {
				classes = append(classes, "animate-pulse")
			}
			div := composition.Div().Classes(classes...)
			dx.AppendChild(div.Element())
		}
	}

	drawSpark := func(v int) {
		if !sparkEl.Truthy() {
			return
		}
		sparkEl.SetHTML("")
		for i := 0; i < v; i++ {
			div := composition.Div().
				Classes("w-4", "h-4", "m-auto", "rounded-full", "outlined").
				Styles(
					"background", "linear-gradient(90deg, #972b2b, #6e347e)",
					"opacity", "0.7",
					"display", "inline-block",
				)
			sparkEl.AppendChild(div.Element())
		}
	}

	renderCart := func() {
		if !cartEl.Truthy() {
			return
		}
		items, _ := cart.Get("items").([]string)
		cartEl.SetHTML("")
		for _, item := range items {
			div := composition.Div().
				Classes("outlined-xl", "border", "border-zinc-200", "dark:border-zinc-700", "p-3", "bg-white/70", "dark:bg-[#111111]").
				Text(item)
			cartEl.AppendChild(div.Element())
		}
	}
	cart.OnChange("items", func(_ any) { renderCart() })
	comp.SetOnMount(func(*core.HTMLComponent) {
		sparkEl = comp.GetRef("spark")
		cartEl = comp.GetRef("cart")
		renderCart()
		renderDx()
		state.Effect(func() func() {
			drawSpark(count.Get())
			return nil
		})
		highlight.HighlightAll()
	})

	addItem := state.Action(func(ctx state.Context) error {
		items, _ := cart.Get("items").([]string)
		id := fmt.Sprintf("Item #%d", len(items)+1)
		items = append([]string{id}, items...)
		cart.Set("items", items)
		return nil
	})
	addHandler := state.UseAction(context.Background(), addItem)
	dom.RegisterHandlerFunc("add", func() { _ = addHandler() })
	dom.RegisterHandlerFunc("undo", func() { cart.Undo() })
	dom.RegisterHandlerFunc("redo", func() { cart.Redo() })
	dom.RegisterHandlerFunc("inc", func() { count.Set(count.Get() + 1) })
	dom.RegisterHandlerFunc("dec", func() {
		v := count.Get()
		if v > 0 {
			count.Set(v - 1)
		}
	})

	dom.RegisterHandlerFunc("showRTML", func() { show("rtml") })
	dom.RegisterHandlerFunc("showGO", func() { show("go") })
	dom.RegisterHandlerFunc("showPreview", func() { show("preview") })

	return c
}

func show(target string) {
	tabs := []string{"rtml", "go", "preview"}
	for _, t := range tabs {
		body := dom.Doc().ByID("body-" + t)
		tab := dom.Doc().ByID("tab-" + t)
		if t == target {
			body.RemoveClass("hidden")
			tab.SetAttr("class", "inline-flex items-center gap-2 px-3 py-2 text-xs font-medium rounded-t-md border-b-2 border-red-500 text-red-600 bg-white/80 dark:bg-[#111111]")
		} else {
			body.AddClass("hidden")
			tab.SetAttr("class", "inline-flex items-center gap-2 px-3 py-2 text-xs font-medium rounded-t-md border-b-2 border-transparent text-zinc-600 dark:text-zinc-300 hover:text-zinc-900 dark:hover:text-zinc-100 hover:bg-white/50 dark:hover:bg-[#141414]")
		}
	}
}
