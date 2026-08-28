# Render tracing

rfw can expose render scheduling and commit timing to browser tooling. Tracing
is disabled by default and retains no records: an application or benchmark owns
the listener and any buffer it needs.

Enable it before the WebAssembly loader starts:

```html
<script>
  window.RFW_RENDER_TRACE = true;
  window.addEventListener("rfw:render-trace", (event) => {
    console.debug(event.detail);
  });
</script>
```

Setting the flag after the Go runtime has started has no effect. This keeps the
disabled path to a cached boolean check and avoids accidentally enabling
production tracing from application state.

## Event contract

Every record is dispatched synchronously as a `CustomEvent` named
`rfw:render-trace`. Schema version 1 has these common fields:

- `schemaVersion`, `sequence`, and the monotonic `timestampMs`;
- `event`, `batchId`, and `renderId`;
- `componentId`, `componentName`, optional `parentComponentId`, and `depth`;
- `cause` and the accumulated `causes` for a coalesced render.

Scheduler records use `scheduled`, `coalesced`, `started`, `committed`,
`cancelled`, or `failed`. Component cleanup emits `unmounted`. Direct
`MountRoot` and router outlet commits emit `started` and a terminal record
without inventing a scheduler queue operation.

A store cause contains `module`, `store`, and `key`. Other cause kinds in the
initial schema are `mount`, `signal`, `router`, `parent`, and `explicit`.
Scheduled records also report `queueDepth` and `coalescedCount`.

Successful terminal records report `templateMs`, `domMs`, `totalMs`, and an
`outcome`. A cancellation or failure carries a `reason` when rfw knows it.

## Measurement boundaries

`templateMs` covers component RTML evaluation. `domMs` covers the matching DOM
update call, including reconciliation and binding refresh. `totalMs` covers the
whole traced render phase. Browser style, layout, paint, garbage collection,
and unrelated JavaScript can extend beyond that interval; measure them with the
Performance or DevTools protocols rather than assigning them to rfw's DOM
counter.

The event transport is intentionally independent of a benchmark framework. A
consumer can install its listener before boot, keep a bounded ring buffer, and
correlate records with its own navigation or interaction markers by monotonic
timestamp.
