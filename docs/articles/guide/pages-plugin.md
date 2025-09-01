# Pages Plugin

The `pages` plugin provides file-based routing by scanning a `pages/`
directory and registering routes for each Go file it finds.

## How It Works

During `PreBuild` the plugin walks the `pages/` folder and generates a
temporary `routes_gen.go` with calls to `router.RegisterRoute`. Each file name
maps to a URL path, and dynamic segments use square brackets to become route
parameters. After the build the generated file is removed.

To activate these registrations, import the generated package, usually with
a blank import in `main.go`:

```go
import _ "your/module/pages"
```

## Usage

Create a `pages/` directory in your module. Each Go file should export a
constructor whose name is the PascalCase version of its path:

```
pages/
  index.go        // -> /
  about.go        // -> /about
  posts/[id].go   // -> /posts/:id
```

`about.go` must provide `func About() core.Component` and so on.
See [`docs/examples/pages`](../../examples/pages) for a complete example.

To customize the directory, set `dir` in the `pages` plugin configuration
inside `rfw.json`.

## Best Practices

- Keep page components focused on layout and data fetching. Reusable pieces
  belong in regular components under `components/`.
- Prefer descriptive file and folder names; the generated route paths mirror
  the directory structure.
- Use dynamic segments like `[id].go` sparingly. If your route logic depends
  on runtime data, register it programmatically with the router API instead.
- Commit only your source files. The generated `routes_gen.go` is temporary
  and should not be checked in.

## When to Use

Use the pages plugin when you want convention-over-configuration routing and
simple page-level components. For applications requiring runtime route
registration or highly dynamic behaviour, use the router API directly instead
of the file-based approach.

