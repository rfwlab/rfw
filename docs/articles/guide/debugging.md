# Debugging

**rfw** includes a development overlay for inspecting components, performance, and logs. Start the dev server with `--debug` to enable it:

```bash
rfw dev --debug
```

In this mode the `bundler` plugin is disabled—JavaScript, CSS, and HTML remain unminified. Use `rfw build` for optimized assets.

## Overlay Features

Toggle the overlay with the floating button (bottom-right) or press `Ctrl+Shift+D`.

* **Components**: view the active component tree
* **Logs**: live console feed with filtering and clearing

  * KPI bar shows FPS, memory usage, node count, and render time
  * Keeps the most recent 200 entries
  * Suppresses logs containing `mutation:` for performance
* **Vars**: searchable tree of `/debug/vars`
* **Pprof**: profiles from `/debug/pprof/`

  * Inline viewer for text profiles
  * Download links for binary profiles

The overlay itself is marked with `data-rfw-ignore` so application mutation observers skip it.

## Notes

* The overlay is only injected in debug mode and excluded from production
* Debug endpoints remain available at `/debug/vars` and `/debug/pprof/` even outside the overlay
