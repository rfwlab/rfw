# WASM Loader

The default project template ships with a Go-powered loader that exposes a global `WasmLoader` helper. The loader displays a red progress bar while the WebAssembly bundle downloads and initializes.

## Why a loader?

Even though WASM files are optimized for delivery, large applications can still take noticeable time to load. The loader provides immediate visual feedback, letting users know that the application is starting up.

## Customizing the loader

Include the bundled `wasm_loader.js` script and invoke `WasmLoader.load` with your `Go` runtime and styling options:

```html
<script src="/wasm_loader.js"></script>
<script>
  const go = new Go();
  WasmLoader.load("/app.wasm", {
    go,
    color: "#ff0000", // bar color
    height: "4px",    // bar height
    blur: "4px",      // glow blur radius
  });
</script>
```

## Using your own loader

If you prefer to implement a different loading experience, set `skipLoader` to `true` and provide your own UI while waiting for the module:

```html
<script src="/wasm_loader.js"></script>
<script>
  const go = new Go();
  WasmLoader.load("/app.wasm", { go, skipLoader: true }).then(() => {
    // custom loader logic here
  });
</script>
```

Alternatively, you can remove the loader script entirely and handle WASM loading manually.
