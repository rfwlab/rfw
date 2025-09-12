# Copy Plugin

## Context

The `copy` plugin moves arbitrary files into the final build output using glob patterns.
It helps bundle example components or other assets that do not fit the default directory layout.

## When to use

Use this plugin when you need to copy specific files or directories into `build/static` or another destination during `rfw build`.
It supports `**` wildcards to match nested paths.

## How to use

Add a configuration block under `plugins.copy` in `rfw.json`:

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

## API

- `from`: source glob pattern. Uses `doublestar` syntax so `**` matches nested directories.
- `to`: destination directory where matched files are placed.

## Example

Suppose your project keeps reusable components under `examples/components`.
The configuration above copies every file, including the `templates` subdirectory, into `build/static/examples/components`.
After running `rfw build`, all matched files appear under the destination preserving their relative structure.

## Limitations

- Only files are copied; directories are created as needed but empty folders are ignored.
- Patterns are evaluated relative to the project root and follow `doublestar` rules.

## Related links

- [Manifest](manifest)
- [Assets Plugin](assets-plugin)
