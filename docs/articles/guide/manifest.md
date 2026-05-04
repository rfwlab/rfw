# Manifest

The project manifest (`rfw.json`) defines build options and plugin configuration. Running `rfw init` generates this file with `build.type` set to `ssc` by default.

## build.type

Controls the build mode:

* **unset** → standard Wasm build
* **ssc** → enables Server-Side Computed builds and compiles host components

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
      "input": "input.css",
      "output": "tailwind.css",
      "minify": true
    }
  }
}
```

* `input`: source file with `@import "tailwindcss"` directives (Tailwind v4)
* `output`: compiled CSS file (default: `tailwind.css`)
* `minify`: set `false` to disable minification
* `args`: extra CLI arguments for `tailwindcss`

### Environment Variables

Expose variables prefixed with `RFW_` through the generated `rfwenv` package.

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
        { "from": "examples/**/*", "to": "build/static/examples" }
      ]
    }
  }
}
```

Plugins run during `rfw build` and also watch files in development.