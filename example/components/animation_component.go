//go:build js && wasm

package components

import (
	_ "embed"
	"time"

	anim "github.com/rfwlab/rfw/v1/animation"
	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
)

//go:embed templates/animation_component.rtml
var animationComponentTpl []byte

func NewAnimationComponent() *core.HTMLComponent {
	c := core.NewComponent("AnimationComponent", animationComponentTpl, nil)

	headerComponent := NewHeaderComponent(map[string]any{
		"title": "Animations",
	})
	c.AddDependency("header", headerComponent)

	dom.RegisterHandlerFunc("animateTranslate", func() {
		anim.Translate("#translateBox", 0, 0, 100, 0, 500*time.Millisecond)
	})
	dom.RegisterHandlerFunc("animateFade", func() {
		anim.Fade("#fadeBox", 1, 0, 500*time.Millisecond)
	})
	dom.RegisterHandlerFunc("animateScale", func() {
		anim.Scale("#scaleBox", 1, 1.5, 500*time.Millisecond)
	})
	dom.RegisterHandlerFunc("animateRainbow", func() {
		colors := []string{"red", "orange", "yellow", "green", "blue", "indigo", "violet"}
		anim.ColorCycle("#rainbowBox", colors, 700*time.Millisecond)
	})

	return c
}
