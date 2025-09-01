# Provide & Inject

Components can share arbitrary values with their descendants without passing them through props. A parent calls `Provide` to expose a value, and any nested component can retrieve it with `Inject`.

## Providing a value

```go
func (c *ParentComponent) OnMount() {
    c.Provide("user", "Ada")
}
```

The `OnMount` hook ensures the value is provided after the component is attached. For more on lifecycle hooks, see the [API reference](../api/core#lifecycle-hooks).

## Consuming a value

`Inject` searches up the component tree until it finds a matching key. Use the generic helper to retrieve a typed value.

```go
name, ok := core.Inject[string](c, "user")
if ok {
    // use name
}
```

## Example

```go
parent := core.NewComponent("Parent", parentTpl, nil)
child := core.NewComponent("Child", childTpl, nil)

parent.Provide("answer", 42)
parent.AddDependency("child", child)

answer, _ := core.Inject[int](child, "answer") // 42
```

This mechanism avoids verbose prop chains while keeping data flow explicit.
