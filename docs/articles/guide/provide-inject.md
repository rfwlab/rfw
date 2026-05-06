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

---

## Injecting a Value

Use `Inject` to search up the component tree for a matching key:

```go
name, ok := core.Inject[string](c, "user")
if ok {
    // use name
}
```

If no provider is found, `ok` is `false`.

---

## DI Container Injection

rfw v2 supports dependency injection via `*t.Inject[T]` struct fields and the `composition.Container()`:

```go
// Register a dependency globally
composition.Container().Provide("logger", myLogger)

// Inject into a struct field
type MyPage struct {
    composition.Component
    Logger *t.Inject[Logger]
}
```

When `composition.New(&MyPage{})` is called, `*t.Inject[T]` fields are automatically resolved from the container using the lowercase field name as the key. The `Logger` field above resolves from `Container().Get("logger")`.

Custom keys work by renaming the field:

```go
type Page struct {
    composition.Component
    AppLog *t.Inject[Logger] // resolves from Container().Get("applog")
}
```

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

## When to Use

* **Provide/Inject** for values that flow down the component tree (parent → child).
* **`*t.Inject[T]` field** for global singletons and services registered in the container.