# Cinema

## Why
The Cinema builder orchestrates keyframes, video, and audio to produce rich presentations. The [Cinema API](../api/cinema) provides chainable methods for scenes and media.

```go
cin := animation.NewCinemaBuilder("#root").
    AddScene("#box", map[string]any{"duration": 500}).
    AddKeyFrame(animation.NewKeyFrame().Add("opacity", "0"), 0).
    AddKeyFrame(animation.NewKeyFrame().Add("opacity", "1"), 1)
```

## When to Use
Use Cinema when several elements must animate in sequence or when video and audio playback need to stay in sync.

```go
cin.PlayVideo().AddPause(time.Second).PlayAudio()
```

## When Not to Use
Skip Cinema for single animations that CSS or the basic [Animation API](../api/animation) can handle.

```css
.fade-in { transition: opacity .5s; }
```

## Interactive Demo
@include:ExampleFrame:{code:"/examples/components/cinema_component.go", uri:"/examples/cinema"}

## Notes and Limitations
- Tracks only one audio element; call `AddAudio` again for different sounds.
- Requires browser environments with video and audio support.

## Related
- [Animation API](../api/animation)
- [DOM API](../api/dom)
