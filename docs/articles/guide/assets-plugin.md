# Assets Plugin

The `assets` plugin copies static files into the build output so they can be served with your application.

## Configuration

Configure the plugin inside `rfw.json` under `plugins.assets`:

```json
{
  "plugins": {
    "assets": {
      "dir": "assets",
      "dest": "dist"
    }
  }
}
```

- `dir`: source directory containing static files (defaults to `assets`).
- `dest`: destination directory inside the build output (`dist` by default).

During `rfw build` the plugin walks the source directory and copies each file into the destination. Changes inside `dir` trigger a rebuild when running `rfw dev`.
