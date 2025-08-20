# cinema

An advanced animation builder for orchestrating scenes, keyframes, and media playback.

- `NewCinemaBuilder(sel)` creates a builder rooted at a DOM element.
- `AddScene(sel, opts)` sets up a new scene with keyframes and timing.
- `AddKeyFrame(frame, offset)` appends a keyframe to the current scene.
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
- `AddScripted(fn)` runs custom scripts.
- `OnStart(fn)` and `OnEnd(fn)` register lifecycle callbacks.

## Usage

Build sequences with keyframes and media controls for rich presentations.

@include:ExampleFrame:{code:"/examples/components/cinema_component.go", uri:"/examples/cinema"}

