# Contributing

## Code Requirements
- Follow Go conventions and run `gofmt` on changed files.
- Keep changes focused; separate unrelated edits into different commits.
- **DO NOT** use `js.Global` directly, use the rfw APIs or, in case of emergency, crash the glass pane and use the `v1/js` APIs (better if you implement the missing ones, instead, but still better than using `js.Global` directly).

## Testing
- Ensure existing tests pass and add tests for new code.
- Run the full suite with:
  ```bash
  go test -v ./...
  ```

## Pull Request Process
- Discuss major changes in an issue before opening a PR.
- Update documentation and changelog when relevant.
- After pushing, fill out the PR template and link to any related issues.
