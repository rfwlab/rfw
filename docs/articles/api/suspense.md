# suspense

Render a fallback while asynchronous content resolves.

| Item | Description |
| --- | --- |
| `Suspense` | Component that shows fallback content until rendering succeeds. |
| `NewSuspense(render func() (string, error), fallback string) *Suspense` | Constructor for `Suspense`. |

