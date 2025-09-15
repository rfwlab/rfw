# Assets

The **assets** package makes it easy to load images, JSON, and binary files at runtime. Loaders are asynchronous and cache results by URL, so repeated requests are instant. See the [API reference](../api/assets) for full details.

## When to Use

Use these loaders when your app needs external data or media after the initial page load:

* Images
* Configuration JSON
* Binary formats such as glTF

Loaders integrate with the [`http` package](../api/http) and return `http.ErrPending` while the request is in progress. This makes them suitable for components that display loading states.

## How It Works

1. Import the package: `import "github.com/rfwlab/rfw/v1/assets"`
2. Call `assets.LoadImage`, `assets.LoadJSON`, or `assets.LoadModel` with a URL
3. Handle `http.ErrPending` to show a loading state
4. Subsequent calls with the same URL return instantly from cache

`LoadJSON` internally uses [`http.FetchJSON`](../api/http#usage) to fetch and decode.

## Example: Load an Image

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

To preload assets, call the loader in `OnMount` so later requests are cached and instant.

## Notes

* Only works in WebAssembly builds
* `LoadModel` returns raw bytes; parsing is up to you
* Clear caches manually with `assets.ClearCache`
* Remote URLs may be blocked by CORS—prefer same-origin files

## Related

* [Assets Plugin](./assets-plugin)
* [http package](../api/http)
* [Assets API](../api/assets)
