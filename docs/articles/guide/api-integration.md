# API Integration

## Why
Applications often need to communicate with external services. The [HTTP API](../api/http) provides helpers that cache responses and surface loading states.

```go
var todo struct {
    Title string `json:"title"`
}
if err := http.FetchJSON("https://jsonplaceholder.typicode.com/todos/1", &todo); err != nil {
    if errors.Is(err, http.ErrPending) {
        // display loading UI
    } else {
        // handle fetch error
    }
}
```

## When to Use
Use `http.FetchJSON` when retrieving remote data to populate stores or render components.

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

## When Not to Use
Skip network requests for static data bundled with the application or when offline support is required.

```go
data := loadFromLocalFile()
store.Set("apiData", data)
```

## Interactive Demo
@include:ExampleFrame:{code:"/examples/components/api_integration_component.go", uri:"/examples/api"}
