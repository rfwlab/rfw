//go:build js && wasm

package components

import (
	_ "embed"

	anim "github.com/rfwlab/rfw/v1/animation"
	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
)

//go:embed templates/cinema_component.rtml
var cinemaComponentTpl []byte

func NewCinemaComponent() *core.HTMLComponent {
	c := core.NewComponent("CinemaComponent", cinemaComponentTpl, nil)
	dom.RegisterHandlerFunc("runCinema", func() {
		anim.NewCinemaBuilder("#cinemaRoot").
			AddScene("#sceneBox", map[string]any{"duration": 1000}).
			AddKeyFrame(map[string]any{"transform": "translateX(0px)"}, 0).
			AddKeyFrame(map[string]any{"transform": "translateX(100px)"}, 1).
			AddVideo("#sceneVideo").
			PlayVideo().
			Play()
	})
	return c
}
