# WASM Loader

The default project template includes a Go-powered loader that exposes a global `WasmLoader` helper. By default, it shows a red progress bar while the WebAssembly bundle downloads and initializes.

> **Note:** When the build pipeline emits a Brotli-compressed bundle (e.g. `app.wasm.br`), the loader automatically tries the `.wasm.br` asset before falling back to the plain `.wasm` file, so existing calls don’t need to change.

---

## Why a Loader?

WASM files are optimized, but large applications may still take noticeable time to load. A loader provides immediate visual feedback so users know the app is starting.

---

## Customizing the Loader

Add the bundled script and call `WasmLoader.load` with the `Go` runtime and your style options:

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

You can change the color, size, and glow to match your branding.

---

## Using Your Own Loader

Disable the built-in UI with `skipLoader: true` and implement a custom one:

```html
<script src="/wasm_loader.js"></script>
<script>
  const go = new Go();
  WasmLoader.load("/app.wasm", { go, skipLoader: true }).then(() => {
    // custom loader logic here
  });
</script>
```

You can also remove the loader script entirely and handle WASM loading manually if you need full control.

---

The default loader is simple and functional, but it’s flexible enough to be styled or replaced as your application grows.
