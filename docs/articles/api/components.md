# components

The `components` package covers helpers and types for building reactive UI elements.

## Caching

`HTMLComponent` optionally caches the result of `Render()` using a hash of its current props and dependencies. Subsequent calls with the same inputs return the cached markup. When props or dependencies change, the hash key changes and the previous cache entry is invalidated, triggering a fresh render.

## CSS/JS minification
When rendering `.rtml` templates `HTMLComponent.Render` automatically minifies inline `<script>` and `<style>` blocks using the [`tdewolff/minify`](https://github.com/tdewolff/minify) library. This keeps in-template snippets small without external tools.

Example:

```rtml
<root></root>
<style>body { color: red; }</style>
<script>function add ( a , b ){ return a + b ; }</script>
```

Renders as:

```rtml
<root></root><style>body{color:red}</style><script>function add(a,b){return a+b}</script>
```
