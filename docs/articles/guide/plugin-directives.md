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
    doc := dom.Doc()

    // [plugin:soccer.badge] → data-plugin="soccer.badge"
    badges := doc.QueryAll("[data-plugin=\"soccer.badge\"]")
    for i := 0; i < badges.Length(); i++ {
        badges.Index(i).SetText("⚽ Lions FC")
    }

    // @plugin:soccer.log → data-plugin-cmd="soccer.log"
    cmds := doc.QueryAll("[data-plugin-cmd=\"soccer.log\"]")
    for i := 0; i < cmds.Length(); i++ {
        cmds.Index(i).OnClick(func(dom.Event) {
            core.Log().Info("Go Lions!")
        })
    }
})
```

## APIs

* `core.RegisterPluginVar(plugin, name, val)`
* `(*core.App).RegisterTemplate(func(componentID, html string))`
* `dom.Doc()`
* `(dom.Document).QueryAll(selector string) dom.Element`
* `(dom.Element).Length() int`
* `(dom.Element).Index(i int) dom.Element`
* `(dom.Element).SetText(text string)`
* `(dom.Element).OnClick(handler func(dom.Event))`
* `core.Log().Info(format string, v ...any)`

## Example

@include\:ExampleFrame:{code:"/examples/components/plugin\_directives\_component.go", uri:"/examples/plugin-directives"}

## Notes

* Commands render as `data-plugin-cmd` and require manual event wiring
* Constructors `[plugin:]` render as `data-plugin`; plugins decide how to populate them

## Related

* [Plugin API](../api/plugins.md#rtml-directives)
* [Template Syntax](../essentials/template-syntax.md)
