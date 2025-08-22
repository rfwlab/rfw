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

var cinema *anim.CinemaBuilder

func NewCinemaComponent() *core.HTMLComponent {
	c := core.NewComponent("CinemaComponent", cinemaComponentTpl, nil)

	dom.RegisterHandlerFunc("playCinema", func() {
		if cinema == nil {
			cinema = anim.NewCinemaBuilder("#cinemaRoot").
				AddScene("#sceneBox", map[string]any{"duration": 1000}).
				AddKeyFrame(anim.NewKeyFrame().Add("transform", "translateX(0px)"), 0).
				AddKeyFrame(anim.NewKeyFrame().Add("transform", "translateX(100px)"), 1).
				AddVideo("#sceneVideo").
				BindProgress("#progressBar")
		}
		cinema.PlayVideo().Play()
	})

	dom.RegisterHandlerFunc("pauseCinema", func() {
		if cinema != nil {
			cinema.PauseVideo()
		}
	})

	dom.RegisterHandlerFunc("stopCinema", func() {
		if cinema != nil {
			cinema.StopVideo()
		}
	})

	dom.RegisterHandlerFunc("seekCinema", func() {
		if cinema != nil {
			cinema.SeekVideo(5)
		}
	})

	dom.RegisterHandlerFunc("speedCinema", func() {
		if cinema != nil {
			cinema.SetPlaybackRate(1.5)
		}
	})

	dom.RegisterHandlerFunc("volumeCinema", func() {
		if cinema != nil {
			cinema.SetVolume(0.5)
		}
	})

	dom.RegisterHandlerFunc("muteCinema", func() {
		if cinema != nil {
			cinema.MuteVideo()
		}
	})

	dom.RegisterHandlerFunc("unmuteCinema", func() {
		if cinema != nil {
			cinema.UnmuteVideo()
		}
	})

	dom.RegisterHandlerFunc("subtitleCinema", func() {
		if cinema != nil {
			cinema.AddSubtitle("subtitles", "English", "en", "https://www.w3schools.com/html/mov_bbb_subtitles_en.vtt")
		}
	})

	dom.RegisterHandlerFunc("audioCinema", func() {
		if cinema != nil {
			cinema.AddAudio("https://www.w3schools.com/html/horse.mp3")
		}
	})

	return c
}
