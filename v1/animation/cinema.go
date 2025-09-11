//go:build js && wasm

package animation

import (
	"strconv"
	"time"

	events "github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
)

type Scene struct {
	el        js.Value
	keyframes []map[string]any
	options   map[string]any
}

type CinemaBuilder struct {
	root    js.Value
	scenes  []*Scene
	current *Scene
	scripts []func()
	video   js.Value
	audio   js.Value
	loop    int
	start   func()
	end     func()
}

func NewCinemaBuilder(sel string) *CinemaBuilder {
	return &CinemaBuilder{root: query(sel), loop: 1}
}

func (c *CinemaBuilder) AddScene(sel string, opts map[string]any) *CinemaBuilder {
	s := &Scene{el: query(sel), options: opts}
	c.scenes = append(c.scenes, s)
	c.current = s
	return c
}

// AddKeyFrame appends a keyframe built with KeyFrameMap to the current scene.
// For more advanced scenarios a raw map can be supplied using AddKeyFrameMap.
func (c *CinemaBuilder) AddKeyFrame(frame KeyFrameMap, offset float64) *CinemaBuilder {
	if c.current == nil {
		c.AddScene("", nil)
	}
	frame["offset"] = offset
	c.current.keyframes = append(c.current.keyframes, map[string]any(frame))
	return c
}

// AddKeyFrameMap appends a raw map as keyframe to the current scene and is kept
// for backwards compatibility and advanced use cases.
func (c *CinemaBuilder) AddKeyFrameMap(frame map[string]any, offset float64) *CinemaBuilder {
	if c.current == nil {
		c.AddScene("", nil)
	}
	frame["offset"] = offset
	c.current.keyframes = append(c.current.keyframes, frame)
	return c
}

func (c *CinemaBuilder) AddTransition(props map[string]any, duration time.Duration) *CinemaBuilder {
	frames := []map[string]any{
		{"offset": 0},
		{"offset": 1},
	}
	for k, v := range props {
		frames[0][k] = v
		frames[1][k] = v
	}
	c.current.keyframes = append(c.current.keyframes, frames...)
	if c.current.options == nil {
		c.current.options = map[string]any{}
	}
	c.current.options["duration"] = duration.Milliseconds()
	return c
}

func (c *CinemaBuilder) AddSequence(fn func(b *CinemaBuilder)) *CinemaBuilder {
	fn(c)
	return c
}

func (c *CinemaBuilder) AddParallel(fn func(b *CinemaBuilder)) *CinemaBuilder {
	go fn(c)
	return c
}

func (c *CinemaBuilder) AddPause(d time.Duration) *CinemaBuilder {
	c.scripts = append(c.scripts, func() { <-time.After(d) })
	return c
}

func (c *CinemaBuilder) AddLoop(count int) *CinemaBuilder {
	if count <= 0 {
		c.loop = 1
	} else {
		c.loop = count
	}
	return c
}

func (c *CinemaBuilder) AddVideo(sel string) *CinemaBuilder {
	c.video = query(sel)
	return c
}

func (c *CinemaBuilder) PlayVideo() *CinemaBuilder {
	if c.video.Truthy() {
		c.video.Call("play")
	}
	return c
}

func (c *CinemaBuilder) PauseVideo() *CinemaBuilder {
	if c.video.Truthy() {
		c.video.Call("pause")
	}
	return c
}

func (c *CinemaBuilder) StopVideo() *CinemaBuilder {
	if c.video.Truthy() {
		c.video.Call("pause")
		c.video.Set("currentTime", 0)
	}
	return c
}

func (c *CinemaBuilder) SeekVideo(t float64) *CinemaBuilder {
	if c.video.Truthy() {
		c.video.Set("currentTime", t)
	}
	return c
}

func (c *CinemaBuilder) SetPlaybackRate(rate float64) *CinemaBuilder {
	if c.video.Truthy() {
		c.video.Set("playbackRate", rate)
	}
	return c
}

func (c *CinemaBuilder) SetVolume(v float64) *CinemaBuilder {
	if c.video.Truthy() {
		c.video.Set("volume", v)
	}
	return c
}

func (c *CinemaBuilder) MuteVideo() *CinemaBuilder {
	if c.video.Truthy() {
		c.video.Set("muted", true)
	}
	return c
}

func (c *CinemaBuilder) UnmuteVideo() *CinemaBuilder {
	if c.video.Truthy() {
		c.video.Set("muted", false)
	}
	return c
}

func (c *CinemaBuilder) AddSubtitle(kind, label, srcLang, src string) *CinemaBuilder {
	if c.video.Truthy() {
		track := js.Document().Call("createElement", "track")
		track.Set("kind", kind)
		track.Set("label", label)
		track.Set("srclang", srcLang)
		track.Set("src", src)
		c.video.Call("appendChild", track)
	}
	return c
}

func (c *CinemaBuilder) AddAudio(src string) *CinemaBuilder {
	audio := js.Document().Call("createElement", "audio")
	audio.Set("src", src)
	audio.Set("controls", true)
	c.root.Call("appendChild", audio)
	c.audio = audio
	return c
}

func (c *CinemaBuilder) PlayAudio() *CinemaBuilder {
	if c.audio.Truthy() {
		c.audio.Set("currentTime", 0)
		c.audio.Call("play")
	}
	return c
}

func (c *CinemaBuilder) PauseAudio() *CinemaBuilder {
	if c.audio.Truthy() {
		c.audio.Call("pause")
	}
	return c
}

func (c *CinemaBuilder) StopAudio() *CinemaBuilder {
	if c.audio.Truthy() {
		c.audio.Call("pause")
		c.audio.Set("currentTime", 0)
	}
	return c
}

func (c *CinemaBuilder) SetAudioVolume(v float64) *CinemaBuilder {
	if c.audio.Truthy() {
		c.audio.Set("volume", v)
	}
	return c
}

func (c *CinemaBuilder) MuteAudio() *CinemaBuilder {
	if c.audio.Truthy() {
		c.audio.Set("muted", true)
	}
	return c
}

func (c *CinemaBuilder) UnmuteAudio() *CinemaBuilder {
	if c.audio.Truthy() {
		c.audio.Set("muted", false)
	}
	return c
}

func (c *CinemaBuilder) BindProgress(sel string) *CinemaBuilder {
	if c.video.Truthy() {
		bar := query(sel)
		if bar.Truthy() {
			bar.Set("max", 100)

			update := func(js.Value) {
				dur := c.video.Get("duration").Float()
				if dur > 0 {
					cur := c.video.Get("currentTime").Float()
					bar.Set("value", cur/dur*100)
				}
			}
			input := func(js.Value) {
				dur := c.video.Get("duration").Float()
				if dur > 0 {
					valStr := bar.Get("value").String()
					if val, err := strconv.ParseFloat(valStr, 64); err == nil {
						c.video.Set("currentTime", val/100*dur)
					}
				}
			}

			events.OnTimeUpdate(c.video, update)
			events.OnInput(bar, input)
		}
	}
	return c
}

func (c *CinemaBuilder) AddScripted(fn func(el js.Value)) *CinemaBuilder {
	c.scripts = append(c.scripts, func() { fn(c.root) })
	return c
}

func (c *CinemaBuilder) OnStart(fn func()) *CinemaBuilder {
	c.start = fn
	return c
}

func (c *CinemaBuilder) OnEnd(fn func()) *CinemaBuilder {
	c.end = fn
	return c
}

func (c *CinemaBuilder) Play() {
	for i := 0; i < c.loop; i++ {
		if c.start != nil {
			c.start()
		}
		for _, s := range c.scenes {
			if len(s.keyframes) > 0 {
				KeyframesForScene(s)
			}
		}
		for _, fn := range c.scripts {
			fn()
		}
		if c.end != nil {
			c.end()
		}
	}
}

func KeyframesForScene(s *Scene) js.Value {
	// convert frames to []any to ensure proper conversion to JS values
	frames := make([]any, len(s.keyframes))
	for i, f := range s.keyframes {
		frames[i] = f
	}

	// convert option numeric types to float64 for syscall/js compatibility
	opts := make(map[string]any, len(s.options))
	for k, v := range s.options {
		switch n := v.(type) {
		case int:
			opts[k] = float64(n)
		case int32:
			opts[k] = float64(n)
		case int64:
			opts[k] = float64(n)
		default:
			opts[k] = v
		}
	}

	return s.el.Call("animate", frames, opts)
}
