# assets

Load images, JSON data, and binary models at runtime with caching and suspense-style loading.

| Function | Description |
| --- | --- |
| `LoadImage(url string) (js.Value, error)` | Asynchronously fetches an image and caches it by URL. |
| `LoadJSON(url string, v any) error` | Fetches JSON into `v` using `http.FetchJSON`. |
| `LoadModel(url string) ([]byte, error)` | Downloads binary data such as glTF models, caching results. |
| `ClearCache(url string)` | Removes a cached entry for `url`. |

