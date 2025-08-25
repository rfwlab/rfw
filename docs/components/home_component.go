//go:build js && wasm

package components

import (
	_ "embed"
	"time"

	anim "github.com/rfwlab/rfw/v1/animation"
	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
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

	// Update the story card using the template markup
	renderSnippet := func(index int) {
		if index < 0 || index >= len(snippets) {
			return
		}

		snippet := snippets[index]
		image := dom.ByID("story-image")
		bars := dom.QueryAll("#progress-bars > div > div")
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

	return c
}
