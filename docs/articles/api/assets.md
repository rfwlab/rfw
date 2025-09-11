# assets

Load images, JSON data, and binary models at runtime with caching and suspense-style loading.

## Context

Applications often need to pull in external resources after the initial page load. The `assets` package provides non-blocking loaders so components remain responsive while files download.

## When to Use

Use these helpers whenever your component requires images, configuration JSON, or binary model data that isn't embedded in the initial bundle.

## How

1. Import the package: `import "github.com/rfwlab/rfw/v1/assets"`.
2. Call a loader with the asset URL.
3. Handle `http.ErrPending` while the request completes.

## API

| Function | Description |
| --- | --- |
| `LoadImage(url string) (js.Value, error)` | Asynchronously fetches an image and caches it by URL. |
| `LoadJSON(url string, v any) error` | Fetches JSON into `v` using [`http.FetchJSON`](http.md#usage). |
| `LoadModel(url string) ([]byte, error)` | Downloads binary data such as glTF models, caching results. |
| `ClearCache(url string)` | Removes a cached entry for `url`. |

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

## Notes and Limitations

- Only available in WebAssembly builds.
- `LoadModel` returns raw bytes; parsing formats like glTF is left to the caller.
- Clearing the cache is manual via `assets.ClearCache`.
- Remote URLs may be blocked by CORS; prefer serving images from the same origin.

## Related

- [Assets guide](../guide/assets.md)
- [http package](http.md)
