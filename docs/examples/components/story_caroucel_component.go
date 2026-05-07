//go:build js && wasm

package components

import (
	_ "embed"
	"time"

	core "github.com/rfwlab/rfw/v2/core"
)

//go:embed templates/story_carousel_component.rtml
var storyCarouselTpl []byte

type StoryCarouselComponent struct {
	*core.HTMLComponent
	snippets []map[string]string
	stop     chan struct{}
}

func NewStoryCarouselComponent(snippets []map[string]string) *core.HTMLComponent {
	c := &StoryCarouselComponent{snippets: snippets}
	c.HTMLComponent = core.NewComponentWith("StoryCarouselComponent", storyCarouselTpl, map[string]any{
		"snippets": snippets,
		"current":  0,
	}, c)
	return c.HTMLComponent
}

func (c *StoryCarouselComponent) OnMount() {
	if len(c.snippets) == 0 {
		return
	}
	c.stop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.stop:
				return
			case <-ticker.C:
				if val, ok := c.Store.Get("current").(int); ok {
					next := (val + 1) % len(c.snippets)
					c.Store.Set("current", next)
				}
			}
		}
	}()
}

func (c *StoryCarouselComponent) OnUnmount() {
	if c.stop != nil {
		close(c.stop)
		c.stop = nil
	}
}
