//go:build js && wasm

package components

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	anim "github.com/rfwlab/rfw/v1/animation"
	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
	js "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/home_component.rtml
var homeTpl []byte

func NewHomeComponent() *core.HTMLComponent {
	count := state.NewSignal(0)
	c := core.NewComponent("HomeComponent", homeTpl, map[string]any{"count": count})

	snippets := []map[string]string{
		{"title": "GO component 1", "image": "/slide-1.png"},
		{"title": "Go component 2", "image": "/slide-2.png"},
		{"title": "RTML", "image": "/slide-3.png"},
		{"title": "Browser Preview", "image": "/slide-4.png"},
	}

	// Track viewed snippets
	viewed := make([]bool, len(snippets))

	// Update the story card using the template markup
	renderSnippet := func(index int) {
		if index < 0 || index >= len(snippets) {
			return
		}

		doc := dom.Doc()
		snippet := snippets[index]
		image := doc.ByID("story-image")
		bars := doc.QueryAll("#progress-bars > div > div")
		if !image.Truthy() || bars.Length() != len(snippets) {
			return
		}

		image.Set("src", snippet["image"])
		image.Set("alt", snippet["title"])

		for i := 0; i < bars.Length(); i++ {
			bar := bars.Index(i)
			classList := bar.Get("classList")
			classList.Call("remove", "bg-white/20", "bg-white/80", "animate-[story-progress_5s_linear]")
			if viewed[i] {
				classList.Call("add", "bg-white/80")
			} else {
				classList.Call("add", "bg-white/20")
			}
			if i == index {
				classList.Call("remove", "bg-white/20")
				classList.Call("add", "bg-white/80", "animate-[story-progress_5s_linear]")
			}
		}

		anim.Fade("#story-image", 0, 1, 500*time.Millisecond)
	}

	// Start the carousel
	go func() {
		current := 0
		for {
			renderSnippet(current)
			viewed[current] = true
			current = (current + 1) % len(snippets)
			if current == 0 {
				for i := range viewed {
					viewed[i] = false
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

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
		if h := js.Get("rfwHighlightAll"); h.Truthy() {
			h.Invoke()
		}
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

	return c
}
