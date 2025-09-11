# cinema

An advanced animation builder for orchestrating scenes, keyframes, and media playback.

- `NewCinemaBuilder(sel)` creates a builder rooted at a DOM element.
- `AddScene(sel, opts)` sets up a new scene with keyframes and timing.
- `AddKeyFrame(frame, offset)` appends a keyframe to the current scene using a `KeyFrameMap`.
- `AddKeyFrameMap(frame, offset)` appends a raw map for advanced cases.
- `NewKeyFrame()` creates an empty `KeyFrameMap` with chainable `Add` and `Delete` helpers.
- `AddTransition(props, duration)` generates keyframes for property transitions.
- `AddSequence(fn)` runs a sequence of builder operations.
- `AddParallel(fn)` executes builder operations in a goroutine.
- `AddPause(d)` inserts pauses between steps.
- `AddLoop(count)` repeats the whole script.
- `AddVideo(sel)` binds a video element for playback control.
- `PlayVideo()` starts the bound video.
- `PauseVideo()` pauses the video.
- `StopVideo()` stops and rewinds the video.
- `SeekVideo(t)` jumps to a timestamp.
- `SetPlaybackRate(rate)` adjusts video speed.
- `SetVolume(v)` updates volume.
- `MuteVideo()` silences the audio.
- `UnmuteVideo()` restores audio.
- `AddSubtitle(kind, label, srcLang, src)` appends a subtitle track.
- `AddAudio(src)` injects an audio element.
- `PlayAudio()` starts the bound audio from the beginning.
- `PauseAudio()` pauses the audio.
- `StopAudio()` pauses and rewinds the audio.
- `SetAudioVolume(v)` updates audio volume.
- `MuteAudio()` silences the audio.
- `UnmuteAudio()` restores audio.
- `AddScripted(fn)` runs custom scripts.
- `OnStart(fn)` and `OnEnd(fn)` register lifecycle callbacks.

## Usage

Build sequences with keyframes and media controls for rich presentations.

@include:ExampleFrame:{code:"/examples/components/cinema_component.go", uri:"/examples/cinema"}

## Audio playback

### Why
Supplement animations with sound effects or music cues.

### When
Use for lightweight audio playback without building a full Web Audio graph.

### How
1. Call `AddAudio(src)` to attach an audio element.
2. Trigger `PlayAudio()` when a sound should play.
3. Adjust output via `SetAudioVolume`, `MuteAudio` or `UnmuteAudio`.

### Example: RTS unit selection

```go
cinema := animation.NewCinemaBuilder("#root").AddAudio("/sounds/select.mp3")
dom.RegisterHandlerFunc("selectUnit", func() {
        cinema.PlayAudio()
})
```

### Limitations
- only a single audio element is tracked; call `AddAudio` with a new source for different sounds

### Related
[animation](animation), [dom](dom)

