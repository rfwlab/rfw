# Cinema

The **Cinema builder** orchestrates complex sequences of keyframes, video, and audio. It is powered by the [Cinema API](../api/cinema), which exposes chainable methods to define scenes and media playback.

## Example: Fade In

```go
cin := animation.NewCinemaBuilder("#root").
    AddScene("#box", map[string]any{"duration": 500}).
    AddKeyFrame(animation.NewKeyFrame().Add("opacity", "0"), 0).
    AddKeyFrame(animation.NewKeyFrame().Add("opacity", "1"), 1)
```

This scene fades the element `#box` from transparent to visible.

## Coordinating Media

Chain video, audio, and pauses to create synced experiences:

```go
cin.PlayVideo().
    AddPause(time.Second).
    PlayAudio()
```

## When to Use

* Animate multiple elements in sequence
* Keep video and audio playback in sync
* Build rich presentations or interactive tutorials

## When to Prefer Simpler Tools

Use CSS or the [Animation API](../api/animation) for simple, single effects:

```css
.fade-in { transition: opacity .5s; }
```

## Interactive Example

@include\:ExampleFrame:{code:"/examples/components/cinema\_component.go", uri:"/examples/cinema"}

## Notes

* Only one audio element is tracked at a time; call `AddAudio` again for different sounds.
* Requires a browser with video and audio support.

## Related

* [Animation API](../api/animation)
* [DOM API](../api/dom)
