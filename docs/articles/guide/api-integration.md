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

The accompanying example component triggers a fetch to a public API when a button is
clicked, performing the request in a goroutine so the event loop stays responsive,
and displays the received data once available.
It fetches data from an API and renders the result.

@include:ExampleFrame:{code:"/examples/components/api_integration_component.go", uri:"/examples/api"}
