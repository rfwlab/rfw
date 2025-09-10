# Debugging

RFW provides a development overlay to inspect the component tree and view
console logs. Start the development server with the `--debug` flag to
automatically inject the overlay into served pages.

```bash
rfw dev --debug
```

In this mode the `bundler` plugin is disabled, leaving JavaScript, CSS and HTML files unminified. Run `rfw build` for optimized assets when preparing a release.

Use the floating button in the bottom-right corner or press
`Ctrl`+`Shift`+`D` to toggle the overlay. The overlay exposes:

- **Components**: a snapshot of the active component hierarchy.
- **Logs**: a real-time console log feed with filtering and clearing.
  Recent FPS, memory usage, node count and render time are displayed in the KPI bar.
- **Vars**: a searchable tree of the data served by `/debug/vars`.
- **Pprof**: profile listings from `/debug/pprof/` with an inline viewer for text profiles and download links for binary ones.
Only the most recent 200 entries are retained and logs containing `mutation:`
are suppressed to keep the page responsive. The overlay itself is marked with
`data-rfw-ignore` so application mutation observers ignore it.

The overlay is only injected in debug mode and is excluded from production
builds.

When running with `--debug`, the same endpoints are available directly in the overlay or by visiting `/debug/vars` and `/debug/pprof/` for deeper analysis.
