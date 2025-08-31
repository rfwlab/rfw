# API Integration

Applications often need to communicate with external services. In **rfw**
you can call APIs using the standard `net/http` package and store the
results in a reactive store:

```go
go func() {
        data, _ := fetchData("https://jsonplaceholder.typicode.com/todos/1")
        store.Set("apiData", data)
}()
```

Remember to use [`http.FetchJSON`](../api/http) for a higher-level
interface that caches results and signals loading states:

```go
var todo Todo
if err := http.FetchJSON("https://jsonplaceholder.typicode.com/todos/1", &todo); err != nil {
        if errors.Is(err, http.ErrPending) {
                // display loading UI
        }
}
```

The accompanying example component triggers a fetch to a public API when a button is
clicked, performing the request in a goroutine so the event loop stays responsive,
and displays the received data once available.
It fetches data from an API and renders the result.

@include:ExampleFrame:{code:"/examples/components/api_integration_component.go", uri:"/examples/api"}
