# wasmloader

```go
import "github.com/rfwlab/rfw/v2/wasmloader"
```

WASM loader with progress bar and automatic brotli decompression for Go WASM apps.

## Load

```go
func Load(url string, opts Options)
```

Loads and instantiates a Go WASM bundle. Supports `.wasm` and `.wasm.br` (brotli compressed) files.

## Options

```go
type Options struct {
    Go         js.Value  // Go runtime instance (required)
    Color      string   // Progress bar color (default: #ff0000)
    Height     string   // Progress bar height (default: 4px)
    Blur      string   // Progress bar blur (default: 8px)
    SkipLoader bool     // Disable progress bar
}
```

## Behavior

1. Tries `.wasm.br` first, then `.wasm`
2. Shows animated progress bar during load
3. Automatically decompresses brotli files
4. Calls `Go.run(instance)` after instantiation

## Example

```go
import "github.com/rfwlab/rfw/v2/wasmloader"

wasmloader.Load("./app.wasm", wasmloader.Options{
    Go:    js.Global().Get("Go"),
    Color: "#00ff00",
})
```