# Plugin Directives

## Context
Plugins can expose data and behavior to templates via RTML. By registering variables, commands and constructors, a plugin lets any component access its features with the `{plugin:}`, `@plugin:` and `[plugin:]` domains.

## Prerequisites
- The plugin is registered with `core.RegisterPlugin`.
- Code is compiled to WebAssembly and runs in a browser environment.

## How to
1. Register plugin values during `Install`:
   ```go
   func (Plugin) Install(a *core.App) {
       core.RegisterPluginVar("soccer", "team", "Lions")
   }
   ```
2. Listen for plugin directives by scanning rendered templates:
   ```go
   a.RegisterTemplate(func(componentID, html string) {
       doc := js.Document()
       // `[plugin:soccer.badge]` becomes `data-plugin="soccer.badge"`
       badges := doc.Call("querySelectorAll", "[data-plugin=\"soccer.badge\"]")
       for i := 0; i < badges.Get("length").Int(); i++ {
           el := badges.Index(i)
           el.Set("textContent", "⚽ Lions FC")
       }
       // `@plugin:soccer.log` becomes `data-plugin-cmd="soccer.log"`
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

## APIs Used
- `core.RegisterPluginVar(plugin, name string, val any)`
- `(*core.App).RegisterTemplate(func(componentID, html string))`
- `js.Document() js.Value`
- `js.FuncOf(func(this js.Value, args []js.Value) any) js.Func`
- `js.Console() js.Value`

## Example
@include:ExampleFrame:{code:"/examples/components/plugin_directives_component.go", uri:"/examples/plugin-directives"}

## Notes and Limitations
- Plugin commands render as `data-plugin-cmd` attributes and must be wired up manually.
- `[plugin:]` constructors render as `data-plugin` attributes; plugins decide how to populate them.

## Related Links
- [Plugin API](../api/plugins.md#rtml-directives)
- [Template Syntax](../essentials/template-syntax.md)
