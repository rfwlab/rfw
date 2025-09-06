# WASM Loader

The default project template ships with a small JavaScript loader that displays a red progress bar while the WebAssembly bundle downloads and initializes.

## Why a loader?

Even though WASM files are optimized for delivery, large applications can still take noticeable time to load. The loader provides immediate visual feedback, letting users know that the application is starting up.

## Customizing the loader

The loader exposes a global `WasmLoader` object with a single `load` method. Pass your `Go` runtime and styling options to customize its appearance:

```html
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
<script>
  const go = new Go();
  WasmLoader.load("/app.wasm", { go, skipLoader: true }).then(() => {
    // custom loader logic here
  });
</script>
```

Alternatively, you can remove the loader script entirely and handle WASM loading manually.
