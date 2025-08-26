# Manifest

The project manifest (`rfw.json`) defines build options and plugin configuration.

## build.type

`build.type` toggles special build modes. Leaving it unset produces a standard Wasm build. Setting it to `ssc` enables Server Side Computed builds and compiles any host components:

```json
{
  "build": {
    "type": "ssc"
  }
}
```

## plugins

The `plugins` section lists build plugins keyed by name. Each plugin accepts its own configuration object.

### Tailwind CSS

Generates a stylesheet using the Tailwind CLI.

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

- `input`: source file containing `@tailwind` directives.
- `output`: compiled CSS file (defaults to `tailwind.css`).
- `minify`: disable minification when set to `false`.
- `args`: additional CLI arguments passed to `tailwindcss`.

### Environment variables

Collects environment variables prefixed with `RFW_` and exposes them through a
generated `rfwenv` package:

```go
import rfwenv "github.com/rfwlab/rfw/docs/rfwenv"

clientID := rfwenv.Get("TWITCH_CLIENT_ID")
```

Provide the variables when invoking `rfw` commands, e.g.
`RFW_TWITCH_CLIENT_ID=abc rfw dev`.

```json
{
  "plugins": {
    "env": {}
  }
}
```

### Static assets

Copies files from a directory into the build output.

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

- `dir`: source directory to copy from (defaults to `assets`).
- `dest`: destination folder inside the build output (`dist` by default).

Plugins run during `rfw build` and may watch relevant files for changes while developing.

