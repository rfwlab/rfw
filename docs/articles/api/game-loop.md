# game loop

Tools for building per-frame update and render loops in Go using the
browser's `requestAnimationFrame`.

## Why

Interactive graphics and games often need predictable frame updates.
This package centralises the loop logic so applications can focus on
state and rendering.

## When to Use

Use the game loop when repeatedly updating state or drawing to the DOM
or canvas. It is unnecessary for one-off animations where the
`animation` helpers suffice.

## How

1. Register update and render callbacks with `OnUpdate` and `OnRender`.
2. Call `Start` to begin scheduling frames.
3. Invoke `Stop` when the loop is no longer needed.

## API

```go
func OnUpdate(func(Ticker))
func OnRender(func(Ticker))
func Start()
func Stop()

type Ticker struct {
        Delta time.Duration
        FPS   float64
}
```

`Ticker` supplies the time elapsed since the previous frame and the
current frame rate.

## Example

```go
package main

import "github.com/rfwlab/rfw/v1/game/loop"

func main() {
        loop.OnUpdate(func(t loop.Ticker) {
                _ = t.Delta // advance physics
        })
        loop.OnRender(func(t loop.Ticker) {
                _ = t.FPS // draw frame
        })
        loop.Start()
}
```

## Notes and Limitations

Frames are scheduled with [`js.RequestAnimationFrame`](./js.md#requestanimationframe)
and timed via `performance.now()`. Callbacks run until `Stop` is invoked.
No throttling beyond the browser's frame pacing is performed.

## Related Links

- [js package](./js.md)
- [animation helpers](./animation.md)
