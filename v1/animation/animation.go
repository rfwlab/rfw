//go:build js && wasm

package animation

import (
	"fmt"
	jst "syscall/js"
	"time"

	js "github.com/rfwlab/rfw/v1/js"
)

// query returns the first element matching sel.
func query(sel string) jst.Value {
	return js.Document().Call("querySelector", sel)
}

// animate drives a requestAnimationFrame loop for the given duration and
// invokes step with the current progress (0..1).
func animate(el jst.Value, duration time.Duration, step func(p float64)) {
	start := js.Performance().Call("now").Float()
	total := float64(duration.Milliseconds())
	var cb jst.Func
	cb = jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		now := js.Performance().Call("now").Float()
		p := (now - start) / total
		if p > 1 {
			p = 1
		}
		step(p)
		if p < 1 {
			js.RequestAnimationFrame(cb)
		} else {
			cb.Release()
		}
		return nil
	})
	js.RequestAnimationFrame(cb)
}

// Translate moves the element selected by sel from the starting coordinates
// to the destination using a translate transform.
func Translate(sel string, fromX, fromY, toX, toY float64, duration time.Duration) {
	el := query(sel)
	animate(el, duration, func(p float64) {
		x := fromX + (toX-fromX)*p
		y := fromY + (toY-fromY)*p
		el.Get("style").Set("transform", fmt.Sprintf("translate(%fpx,%fpx)", x, y))
	})
}

// Fade transitions the element's opacity from 'from' to 'to'.
func Fade(sel string, from, to float64, duration time.Duration) {
	el := query(sel)
	animate(el, duration, func(p float64) {
		val := from + (to-from)*p
		el.Get("style").Set("opacity", val)
	})
}

// Scale scales the element from the starting factor to the ending factor.
func Scale(sel string, from, to float64, duration time.Duration) {
	el := query(sel)
	animate(el, duration, func(p float64) {
		val := from + (to-from)*p
		el.Get("style").Set("transform", fmt.Sprintf("scale(%f)", val))
	})
}

// ColorCycle iterates the element's background color through the provided
// list over the given duration.
func ColorCycle(sel string, colors []string, duration time.Duration) {
	el := query(sel)
	stepDur := duration / time.Duration(len(colors))
	go func() {
		for _, c := range colors {
			el.Get("style").Set("background", c)
			<-time.After(stepDur)
		}
	}()
}
