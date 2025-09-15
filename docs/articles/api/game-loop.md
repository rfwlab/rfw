# game loop

Tools for building per-frame update and render loops in Go using the browser's `requestAnimationFrame`.

| Function | Description |
| --- | --- |
| `OnUpdate(func(Ticker))` | Register an update callback. |
| `OnRender(func(Ticker))` | Register a render callback. |
| `Start()` | Begin scheduling frames. |
| `Stop()` | Halt the loop. |
| `Ticker` | Provides `Delta` and `FPS` for frame timing. |

