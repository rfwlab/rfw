# File-based Routing

The `pages` plugin allows defining routes using the filesystem. Place Go files
under a `pages/` directory and the CLI will automatically register them during
builds.

```
pages/
  index.go        // -> /
  about.go        // -> /about
  posts/[id].go   // -> /posts/:id
```

Each file must provide a constructor named after the PascalCase version of its
path. For example, `about.go` should define `func About() core.Component`.
The plugin generates a temporary `routes_gen.go` with calls to
`router.RegisterRoute` for each page.

To activate these registrations in your app, import the generated package —
usually with a blank import — in your entrypoint:

```
import _ "your/module/pages"
```

The repository includes a working example in
`docs/examples/pages` that defines `index.go`, `about.go` and
`posts/[id].go`, mirroring the structure above and registering the
routes `/`, `/about` and `/posts/:id`. The documentation site itself uses
the `pages` plugin for its home (`/`) and about (`/about`) pages, while the
`docs` plugin continues to serve the documentation content.

Dynamic segments use square brackets and become route parameters. A file named
`posts/[id].go` registers the path `/posts/:id`.
