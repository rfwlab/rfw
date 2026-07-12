# devtools

In-page developer tools overlay for rfw applications. Press
`Ctrl+Shift+D` to toggle a fixed dark panel with two tabs:

- **Components**: tree of live components (name and ID), walked from the
  router's current component through `Dependencies`, plus any other
  components observed via mount/unmount lifecycle hooks.
- **Stores**: every store registered on `state.GlobalStoreManager`, with
  key/values JSON-stringified (truncated to 200 characters) and a Refresh
  button.

The plugin has zero cost when not installed; the overlay DOM is built
lazily on first toggle and styled inline (no external CSS).

## Usage

```go
//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/plugins/devtools"
)

func main() {
	core.RegisterPlugin(devtools.New())
	// ... register components, router, etc.
}
```

The [shortcut](../shortcut) plugin is declared as an optional dependency
and installed automatically so the keybinding works out of the box. To
change the binding or the inspected store manager, set the fields before
registering:

```go
p := devtools.New()
p.Shortcut = "control+shift+i"
core.RegisterPlugin(p)
```

You can also toggle programmatically with `p.Toggle()`, `p.Show()` and
`p.Hide()`.

## Limitations

- core's internal dev component registry (`devRegisterComponent`) is
  unexported and a no-op in js/wasm builds, so components mounted before
  the plugin was installed and not reachable from the current route are
  not listed.
- Store inspection uses `StoreManager.Snapshot()`; signals and computed
  values not backed by a store key are not shown.
