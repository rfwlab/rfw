# Copy Plugin

The **copy plugin** moves arbitrary files into the build output using glob patterns. It’s useful for bundling components, examples, or assets that don’t follow the default layout.

## When to Use

* Copy specific files or directories into `build/static` or another destination
* Include examples or extra resources alongside the main bundle
* Match nested paths with `**` wildcards

## Configuration

Add to `rfw.json` under `plugins.copy`:

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

## Options

* **from**: source glob pattern (`doublestar` syntax, `**` matches nested directories)
* **to**: destination directory inside the build output

## Example

With the config above, every file under `examples/components` (including subfolders like `templates/`) is copied into `build/static/examples/components`. Running `rfw build` preserves the structure at the new destination.

## Limitations

* Only files are copied; empty directories are skipped
* Patterns are relative to the project root

## Related

* [Manifest](manifest)
* [Assets Plugin](assets-plugin)
