# Assets

Efficient asset loading allows applications to fetch images, JSON data, and binary models at runtime without blocking rendering.
The `assets` package provides asynchronous helpers with caching so repeated requests return immediately. See the [assets API reference](../api/assets) for detailed function signatures.

## When to Use

Use these loaders when your application needs to fetch external images, configuration JSON or binary model formats such as glTF
after the initial page load. They integrate with the [`http` package](../api/http) and return `http.ErrPending` while a
request is in flight, making them suitable for components that rely on suspense patterns.

## How It Works

1. Import the package: `import "github.com/rfwlab/rfw/v1/assets"`.
2. Call `assets.LoadImage`, `assets.LoadJSON` or `assets.LoadModel` with a URL.
3. Handle `http.ErrPending` to show a loading state while the asset downloads.
4. The result is cached by URL; subsequent calls return instantly.

All loaders return `http.ErrPending` until the asset is ready. `LoadJSON` delegates to [`http.FetchJSON`](../api/http#usage)
for fetching and decoding.

## Example

```go
var img js.Value
if v, err := assets.LoadImage("/static/logo.png"); err != nil {
    if errors.Is(err, http.ErrPending) {
        return core.Text("loading...")
    }
    return core.Text("failed")
} else {
    img = v
}
return core.Img().Attr("src", img.Get("src").String())
```

To preload assets, invoke the loader during an initialization phase such as a component's `OnMount` hook. Later calls will hit
the cache and avoid additional network requests.

## Notes and Limitations

- Only available in WebAssembly builds.
- `LoadModel` returns raw bytes; parsing formats like glTF is left to the caller.
- Clearing the cache is manual via `assets.ClearCache`.
- Remote URLs may be blocked by CORS; prefer serving images from the same origin.

## Related

- [Assets Plugin](./assets-plugin)
- [http package](../api/http)
- [Assets API](../api/assets)
