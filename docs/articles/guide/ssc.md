# Server Side Computed

The project manifest (`rfw.json`) can declare the build type. Setting it to `ssc` enables Server Side Computed builds.

```json
{
  "build": {
    "type": "ssc"
  }
}
```

When `ssc` is active, `rfw build` compiles the Wasm bundle and also builds the Go sources in the `host` directory into a server binary. The server keeps variables and commands prefixed with `h:` synchronized with the client through a persistent WebSocket connection.
