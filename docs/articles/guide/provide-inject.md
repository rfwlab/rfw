# Provide & Inject

Components can share values with their descendants without passing them explicitly through props. A parent calls `Provide` to expose a value, and nested components can access it with `Inject`.

---

## Providing a Value

Call `Provide` inside a lifecycle hook such as `OnMount` to make a value available to child components:

```go
func (c *ParentComponent) OnMount() {
    c.Provide("user", "Ada")
}
```

The `OnMount` hook ensures the value is provided after the component is attached. See [lifecycle hooks](../api/core#lifecycle-hooks) for details.

---

## Consuming a Value

Use `Inject` to search up the component tree for a matching key. With the generic form, you get a typed result:

```go
name, ok := core.Inject[string](c, "user")
if ok {
    // use name
}
```

If no provider is found, `ok` is `false`.

---

## Full Example

```go
parent := core.NewComponent("Parent", parentTpl, nil)
child := core.NewComponent("Child", childTpl, nil)

parent.Provide("answer", 42)
parent.AddDependency("child", child)

answer, _ := core.Inject[int](child, "answer") // 42
```

---

## Why Use It

* Avoids prop drilling for deeply nested components
* Keeps data flow explicit and local to the component tree
* Useful for cross-cutting concerns like themes, localization, or user context
