# API Integration

Most applications need to communicate with external services. In **rfw**, the [HTTP API](../api/http) provides helpers to fetch JSON, handle loading states, and cache responses.

## Fetching JSON

Use `FetchJSON` to retrieve data and unmarshal it into a Go struct:

```go
var todo struct {
    Title string `json:"title"`
}

if err := http.FetchJSON("https://jsonplaceholder.typicode.com/todos/1", &todo); err != nil {
    if errors.Is(err, http.ErrPending) {
        // show loading UI
    } else {
        // handle fetch error
    }
}
```

If the request is pending, `ErrPending` is returned so you can display a loading state. On success, the struct is populated with the response.

## Updating State from a Request

A common pattern is to fetch asynchronously and update a store:

```go
go func() {
    var todo struct {
        Title string `json:"title"`
    }
    if err := http.FetchJSON("https://jsonplaceholder.typicode.com/todos/1", &todo); err != nil {
        store.Set("apiData", err.Error())
        return
    }
    store.Set("apiData", todo.Title)
}()
```

This updates the `apiData` store with either the title or an error string.

## When Not to Use

Skip network calls when:

* The data is static and bundled in your app.
* You need offline-first behavior.

```go
data := loadFromLocalFile()
store.Set("apiData", data)
```

## Interactive Example

@include\:ExampleFrame:{code:"/examples/components/api\_integration\_component.go", uri:"/examples/api"}
