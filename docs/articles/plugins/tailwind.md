# Tailwind Plugin

The **Tailwind plugin** integrates [Tailwind CSS](https://tailwindcss.com) into rfw projects. It runs the `tailwindcss` CLI during the build process and automatically rebuilds when relevant files change.

## Features

* Runs `tailwindcss` with your chosen input and output files.
* Supports minification with a simple flag.
* Accepts extra CLI arguments for advanced configuration.
* Triggers rebuilds when CSS, RTML, HTML, or Go files are modified.

## Usage

Register the plugin in your app configuration. By default, it looks for `index.css` and generates `tailwind.css`.

```go
import (
    core "github.com/rfwlab/rfw/v1/core"
    tailwind "github.com/rfwlab/rfw/v1/plugins/tailwind"
)

func main() {
    core.RegisterPlugin(&tailwind.plugin{})
}
```

In most projects you won’t need to register manually—`rfw` automatically detects and runs the plugin during build if configured.

### Example Configuration

In `rfw.json`:

```json
{
  "plugins": {
    "tailwind": {
      "input": "src/styles.css",
      "output": "dist/tailwind.css",
      "minify": true
    }
  }
}
```

## API Reference

The Tailwind plugin is configured through JSON in your `rfw.json` plugins section.

| Field    | Type      | Default        | Description                                  |
| -------- | --------- | -------------- | -------------------------------------------- |
| `input`  | string    | `index.css`    | Entry CSS file.                              |
| `output` | string    | `tailwind.css` | Generated CSS file.                          |
| `minify` | bool      | `true`         | Whether to minify the output.                |
| `args`   | \[]string | `[]`           | Extra CLI arguments passed to `tailwindcss`. |

## Notes

* Requires the `tailwindcss` binary installed and available in `PATH`.
* Logs build progress and errors to the console.
* Only rebuilds when files with `.css` (excluding the output file), `.rtml`, `.html`, or `.go` extensions change.
* If `tailwindcss` is missing, the plugin fails with a clear error message.
