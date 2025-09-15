# cinema

An advanced animation builder for orchestrating scenes, keyframes, and media playback.

| Function | Description |
| --- | --- |
| `NewCinemaBuilder(sel)` | Create a builder rooted at a DOM element. |
| `AddScene(sel, opts)` | Set up a new scene with keyframes and timing. |
| `AddKeyFrame(frame, offset)` | Append a keyframe to the current scene using a `KeyFrameMap`. |
| `AddKeyFrameMap(frame, offset)` | Append a raw map for advanced cases. |
| `NewKeyFrame()` | Create an empty `KeyFrameMap` with chainable `Add` and `Delete` helpers. |
| `AddTransition(props, duration)` | Generate keyframes for property transitions. |
| `AddSequence(fn)` | Run a sequence of builder operations. |
| `AddParallel(fn)` | Execute builder operations in a goroutine. |
| `AddPause(d)` | Insert pauses between steps. |
| `AddLoop(count)` | Repeat the whole script. |
| `AddVideo(sel)` | Bind a video element for playback control. |
| `PlayVideo()` | Start the bound video. |
| `PauseVideo()` | Pause the video. |
| `StopVideo()` | Stop and rewind the video. |
| `SeekVideo(t)` | Jump to a timestamp. |
| `SetPlaybackRate(rate)` | Adjust video speed. |
| `SetVolume(v)` | Update volume. |
| `MuteVideo()` | Silence the audio. |
| `UnmuteVideo()` | Restore audio. |
| `AddSubtitle(kind, label, srcLang, src)` | Append a subtitle track. |
| `AddAudio(src)` | Inject an audio element. |
| `PlayAudio()` | Start the bound audio from the beginning. |
| `PauseAudio()` | Pause the audio. |
| `StopAudio()` | Pause and rewind the audio. |
| `SetAudioVolume(v)` | Update audio volume. |
| `MuteAudio()` | Silence the audio. |
| `UnmuteAudio()` | Restore audio. |
| `AddScripted(fn)` | Run custom scripts. |
| `OnStart(fn)` | Register a start callback. |
| `OnEnd(fn)` | Register an end callback. |

