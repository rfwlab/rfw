//go:build js && wasm

package components

import (
	"context"
	_ "embed"
	"fmt"

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
			d := dom.CreateElement("div")
			d.AddClass("h-3")
			d.AddClass("outlined")
			d.AddClass("bg-zinc-200/60")
			d.AddClass("dark:bg-zinc-800")
			if i%7 == 0 {
				d.AddClass("animate-pulse")
			}
			dx.AppendChild(d)
		}
	}

	drawSpark := func(v int) {
		spark := dom.Doc().ByID("spark")
		if !spark.Truthy() {
			return
		}
		spark.SetHTML("")

		for i := 0; i < v; i++ {
			dot := dom.CreateElement("div")
			dot.AddClass("w-4")
			dot.AddClass("h-4")
			dot.AddClass("m-auto")
			dot.AddClass("rounded-full")
			dot.AddClass("outlined")
			dot.SetStyle("background", "linear-gradient(90deg, #972b2b, #6e347e)")
			dot.SetStyle("opacity", "0.7")
			dot.SetStyle("display", "inline-block")
			spark.AppendChild(dot)
		}
	}

	renderCart := func() {
		cartEl := dom.Doc().ByID("cart")
		if !cartEl.Truthy() {
			return
		}
		cartEl.SetHTML("")
		if items, ok := cart.Get("items").([]string); ok {
			for _, item := range items {
				d := dom.CreateElement("div")
				d.AddClass("outlined-xl")
				d.AddClass("border")
				d.AddClass("border-zinc-200")
				d.AddClass("dark:border-zinc-700")
				d.AddClass("p-3")
				d.AddClass("bg-white/70")
				d.AddClass("dark:bg-[#111111]")
				d.SetText(item)
				cartEl.AppendChild(d)
			}
		}
	}
	cart.OnChange("items", func(_ any) { renderCart() })
	c.SetOnMount(func(*core.HTMLComponent) {
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
