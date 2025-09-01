# http

Helpers for making HTTP requests in the browser using the JavaScript `fetch` API.

| Function | Description |
| --- | --- |
| `FetchJSON(url, v)` | Fetches JSON from `url` into `v`. Results are cached by URL and the function returns `ErrPending` while the request is in flight. |
| `ClearCache(url)` | Removes the cached response for `url`. |

`FetchJSON` integrates basic caching and a suspense-style API. Calling it while a request is ongoing returns `ErrPending`, allowing callers to defer rendering until data is ready and to pair naturally with [`Suspense`](core#suspense).

## Usage

```go
var todo struct {
        ID    int    `json:"id"`
        Title string `json:"title"`
}
err := http.FetchJSON("https://jsonplaceholder.typicode.com/todos/1", &todo)
if err != nil {
        if errors.Is(err, http.ErrPending) {
                // show loading state
        } else {
                // handle fetch error
        }
}
```
