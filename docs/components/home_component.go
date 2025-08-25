//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"
	"time"

	"syscall/js"

	"github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/home_component.rtml
var homeTpl []byte

func NewHomeComponent() *core.HTMLComponent {
	c := core.NewComponent("HomeComponent", homeTpl, nil)

	snippets := []map[string]string{
		{"title": "GO component 1", "image": "/slide-1.png"},
		{"title": "Go component 2", "image": "/slide-2.png"},
		{"title": "RTML", "image": "/slide-3.png"},
		{"title": "Browser Preview", "image": "/slide-4.png"},
	}

	// Track viewed snippets
	viewed := make([]bool, len(snippets))

	// Render the snippet into a story-style card
	renderSnippet := func(index int) {
		if index < 0 || index >= len(snippets) {
			return
		}
		snippet := snippets[index]
		progressBars := ""
		for i := range len(snippets) {
			progressBars += fmt.Sprintf(
				`<div class="h-1 rounded-full flex-1 mx-1 bg-white/20">
					<div class="h-1 rounded-full %s %s"></div>
				</div>`,
				func() string {
					if viewed[i] {
						return "bg-white/80"
					}
					return "bg-white/20"
				}(),
				func() string {
					if i == index && !viewed[i] {
						return "bg-white/80 rounded-full animate-[story-progress_5s_linear]"
					}
					return ""
				}(),
			)
		}
		html := fmt.Sprintf(`
				<div class="relative w-full p-3 pt-11 rounded-2xl shadow-2xl text-white overflow-hidden"
						style="background-image: url('/slide-bg.png'); background-size: cover; background-position: center; min-height: 400px; min-width: 400px;">

					<!-- Image block -->
					<img src="%s" alt="%s" class="w-full h-auto rounded-md shadow-lg opacity-0 transition-opacity duration-500" id="story-image" style="min-height: 200px; min-width: 300px;">

					<!-- Progress bar (like stories) -->
					<div class="absolute top-5 left-0 w-4/5 flex mx-auto right-0 h-1">
						%s
					</div>
				</div>
		`, snippet["image"], snippet["title"], progressBars)

		container := js.Global().Get("document").Call("getElementById", "examples")
		container.Set("innerHTML", html)

		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
			img := js.Global().Get("document").Call("getElementById", "story-image")
			if !img.IsNull() {
				img.Get("classList").Call("remove", "opacity-0")
			}
			return nil
		}), 50)
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

	return c
}
