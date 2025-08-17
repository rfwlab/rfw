# RTML Templates

RTML is a lightweight template language that looks like HTML but adds a
few conveniences for reactive Go applications.

- **Bindings** are written with `{expr}` and update when the referenced
  state changes.
- **Directives** such as `on:click` or `if:` attach event handlers and
  conditional rendering logic.
- The template is compiled to Go code so no runtime parser runs in the
  browser.

Because RTML compiles to Go, complex components still benefit from the
compiler's type checking and tooling support.
