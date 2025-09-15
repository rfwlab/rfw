# Manifest

The project manifest (`rfw.json`) defines build options and plugin configuration. Running `rfw init` generates this file with `build.type` set to `ssc` by default.

## build.type

Controls the build mode:

* **unset** → standard Wasm build
* **ssc** → enables Server Side Computed builds and compiles host components

```json
{
  "build": {
    "type": "ssc"
  }
}
```

## plugins

The `plugins` section lists build plugins, each with its own configuration.

### Tailwind CSS

Generate stylesheets using the Tailwind CLI:

```json
{
  "plugins": {
    "tailwind": {
      "input": "static/input.css",
      "output": "static/tailwind.css",
      "minify": true
    }
  }
}
```

* `input`: source file with `@tailwind` directives
* `output`: compiled CSS file (default: `tailwind.css`)
* `minify`: set `false` to disable minification
* `args`: extra CLI arguments for `tailwindcss`

### Environment Variables

Expose variables prefixed with `RFW_` through the generated `rfwenv` package:

```go
import rfwenv "github.com/rfwlab/rfw/docs/rfwenv"

clientID := rfwenv.Get("TWITCH_CLIENT_ID")
```

Provide them when running commands:

```bash
RFW_TWITCH_CLIENT_ID=abc rfw dev
```

```json
{
  "plugins": {
    "env": {}
  }
}
```

### Static Assets

Copy files from a directory into the build:

```json
{
  "plugins": {
    "assets": {
      "dir": "public",
      "dest": "dist"
    }
  }
}
```

* `dir`: source directory (default: `assets`)
* `dest`: output folder (default: `dist`)

### Copy Files

Copy files based on glob patterns:

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

* `from`: source glob pattern (`**` matches nested directories)
* `to`: destination directory

### Documentation Content

Bundle markdown articles and sidebar into the build:

```json
{
  "plugins": {
    "docs": {}
  }
}
```

* `dir`: source docs folder (default: `articles`)
* `dest`: output folder (default: `build/static`)

---

Plugins run during `rfw build` and also watch files in development.
