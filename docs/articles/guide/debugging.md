# Debugging

RFW provides a development overlay to inspect the component tree and view
console logs. Start the development server with the `--debug` flag to
automatically inject the overlay into served pages.

```bash
rfw dev --debug
```

Use the floating button in the bottom-right corner or press
`Ctrl`+`Shift`+`D` to toggle the overlay. The overlay exposes:

- **Components**: a snapshot of the active component hierarchy.
- **Logs**: a real-time console log feed with filtering and clearing.
  Recent FPS, memory usage, node count and render time are displayed in the KPI bar.
Only the most recent 200 entries are retained and logs containing `mutation:`
are suppressed to keep the page responsive. The overlay itself is marked with
`data-rfw-ignore` so application mutation observers ignore it.

The overlay is only injected in debug mode and is excluded from production
builds.

When running with `--debug`, profiling endpoints remain available at
`/debug/vars` and `/debug/pprof/` for deeper analysis.
