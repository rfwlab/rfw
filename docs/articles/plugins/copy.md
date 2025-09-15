# Copy Plugin

The **Copy plugin** copies files into the build output using glob patterns, letting you include resources that don’t follow the default layout.

## Features

* Copy files or directories into `build/static` or another location.
* Preserve folder structure.
* Supports `**` wildcards for nested paths (using `doublestar`).

## Configuration

Enable in `rfw.json`:

```json
{
  "plugins": {
    "copy": {
      "files": [
        { "from": "examples/components/**/*", "to": "build/static/examples/components" }
      ]
    }
  }
}
```

### Options

* **from** – source glob pattern (supports `**` for recursive matches).
* **to** – destination directory inside the build output.

## Example

With the above config, all files under `examples/components` (including nested folders) are copied to `build/static/examples/components`. The structure is preserved in the build output.

Run:

```bash
rfw build
```

## Notes

* Only files are copied; empty directories are skipped.
* Patterns are relative to the project root.
