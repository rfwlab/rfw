# Assets Plugin

The **assets plugin** copies static files into the build output so they can be served with your application.

## Configuration

Add the plugin to `rfw.json` under `plugins.assets`:

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

* **dir**: source directory for static files (default: `assets`)
* **dest**: destination directory inside the build output (default: `dist`)

## How It Works

* On `rfw build`, the plugin walks the source directory and copies files into the destination.
* On `rfw dev`, changes in `dir` trigger a rebuild so static assets stay in sync.

Use this plugin to bundle images, fonts, and other resources alongside your application without extra configuration.
