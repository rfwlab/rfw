# Contributing

Thanks for contributing to rfw. Please read this document before opening a
pull request.

## Development setup

You need Go **1.25** or later (see `go.mod`).

Clone the repository and verify everything builds for both targets:

```bash
# native build (host, CLI, SSC server side)
go build ./...

# WebAssembly build (browser side)
GOOS=js GOARCH=wasm go build ./...
```

Both commands must succeed. Any code path guarded by `js && wasm` build tags
must build for both targets; the CLI (`cmd/rfw`) is excluded from wasm builds,
everything else is expected to compile on both.

Run the test suite with:

```bash
go test -v ./...
```

## Code requirements

- Follow Go conventions and run `gofmt` on changed files. Unformatted code is
  not accepted.
- Keep changes focused; separate unrelated edits into different commits.
- **Do not** use `js.Global` directly. Use the rfw APIs (the `dom` and `js`
  packages). If something is missing, prefer implementing the missing helper
  in those packages over reaching for `js.Global`.
- Ensure existing tests pass and add tests for new code.

## Commit messages

- A single imperative subject line. No long bodies.
- Use a type prefix: `feat:`, `fix:`, `refactor:`, `docs:`, `style:`,
  `build:`, `ci:`, `chore:`.
- When a commit resolves an issue, reference it in the prefix:
  `type[closes #N]: subject`, for example:

  ```
  fix[closes #30]: invalidate render cache on reactive store change
  ```

Examples of good subjects:

```
feat: delegated element handlers and input helpers
fix: pass resolved element to delegated handlers
style: gofmt stub files
```

## Pull request process

- Discuss major changes in an issue before opening a PR. Note that core is
  under a scope freeze until v2.0.0 stable (see `ROADMAP.md`): only fixes,
  stability work, and documentation are accepted in core.
- Update documentation and `CHANGELOG.md` (Unreleased section) when relevant.
  Flag breaking changes as **breaking** with migration notes.
- After pushing, fill out the PR template and link to any related issues.
