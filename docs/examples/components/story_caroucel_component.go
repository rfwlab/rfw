//go:build js && wasm

package components

import (
	_ "embed"
	"time"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/story_carousel_component.rtml
var storyCarouselTpl []byte

func NewStoryCarouselComponent(snippets []map[string]string) *core.HTMLComponent {
	c := core.NewComponent("StoryCarouselComponent", storyCarouselTpl, map[string]any{
		"snippets": snippets,
		"current":  0,
	})

	// Automatically advance the story every 5 seconds
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if val, ok := c.Store.Get("current").(int); ok {
				next := (val + 1) % len(snippets)
				c.Store.Set("current", next)
			}
		}
	}()

	return c
}
