# Plugin Directives

Plugins can expose data and behavior to templates via **RTML directives**. By registering variables, commands, and constructors, a plugin makes its features available with the `{plugin:}`, `@plugin:`, and `[plugin:]` domains.

## Prerequisites

* Plugin registered with `core.RegisterPlugin`
* Compiled to WebAssembly and running in a browser

## Registering Values

Inside a plugin’s `Install` method, register variables:

```go
func (Plugin) Install(a *core.App) {
    core.RegisterPluginVar("soccer", "team", "Lions")
}
```

This makes `{plugin:soccer.team}` available in RTML.

## Listening for Directives

Scan rendered templates to handle plugin directives:

```go
a.RegisterTemplate(func(componentID, html string) {
    doc := js.Document()

    // [plugin:soccer.badge] → data-plugin="soccer.badge"
    badges := doc.Call("querySelectorAll", "[data-plugin=\"soccer.badge\"]")
    for i := 0; i < badges.Get("length").Int(); i++ {
        el := badges.Index(i)
        el.Set("textContent", "⚽ Lions FC")
    }

    // @plugin:soccer.log → data-plugin-cmd="soccer.log"
    cmds := doc.Call("querySelectorAll", "[data-plugin-cmd=\"soccer.log\"]")
    handler := js.FuncOf(func(this js.Value, args []js.Value) any {
        js.Console().Call("log", "Go Lions!")
        return nil
    })
    for i := 0; i < cmds.Get("length").Int(); i++ {
        cmds.Index(i).Call("addEventListener", "click", handler)
    }
})
```

## APIs

* `core.RegisterPluginVar(plugin, name, val)`
* `(*core.App).RegisterTemplate(func(componentID, html string))`
* `js.Document()`
* `js.FuncOf(...)`
* `js.Console()`

## Example

@include\:ExampleFrame:{code:"/examples/components/plugin\_directives\_component.go", uri:"/examples/plugin-directives"}

## Notes

* Commands render as `data-plugin-cmd` and require manual event wiring
* Constructors `[plugin:]` render as `data-plugin`; plugins decide how to populate them

## Related

* [Plugin API](../api/plugins.md#rtml-directives)
* [Template Syntax](../essentials/template-syntax.md)
