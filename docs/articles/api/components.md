# components

The `components` package covers helpers and types for building reactive UI elements.

## Caching

`HTMLComponent` optionally caches the result of `Render()` using a hash of its current props and dependencies. Subsequent calls with the same inputs return the cached markup. When props or dependencies change, the hash key changes and the previous cache entry is invalidated, triggering a fresh render.

