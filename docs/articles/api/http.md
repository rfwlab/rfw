# http

Helpers for making HTTP requests in the browser using the JavaScript `fetch` API.

| Function | Description |
| --- | --- |
| `FetchJSON(url, v)` | Fetches JSON from `url` into `v`. Results are cached by URL and return `ErrPending` while the request is in flight. |
| `FetchText(url) (string, error)` | Fetches text from `url`. Results are cached by URL and return `ErrPending` while the request is in flight. |
| `ClearCache(url)` | Removes the cached response for `url`. |
| `RegisterHTTPHook(fn)` | Adds a callback invoked on request start and completion. |

