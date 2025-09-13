# Selective DOM Patching

rfw updates the page through **Selective DOM Patching**, a minimal diffing
routine that works directly on real DOM nodes. When a component renders, the
framework parses the new markup into a fragment and walks it alongside the
existing subtree:

- elements annotated with `[key {expr}]` are matched and moved if necessary;
- text and attributes are updated in place;
- missing nodes are created and obsolete ones removed.

Because the browser's DOM is the only source of truth, there is no virtual
tree to allocate or reconcile. The patcher simply mutates what is already in
the document.

## Benefits of skipping a virtual DOM

- **Lower memory use** – no second representation of the interface.
- **Smaller runtime** – the patcher fits in a handful of functions.
- **Immediate feedback** – updates hit the real nodes without an extra
  abstraction layer.

## Comparison with other frameworks

### Virtual DOM libraries (React, Vue)
These frameworks build a mirror tree in JavaScript and run a diff algorithm to
produce patches. The approach is flexible but introduces memory churn and an
additional abstraction layer.

### Compile-time DOM generation (Svelte)
Svelte analyzes templates ahead of time and emits imperative DOM operations.
There is no diffing at runtime, but the generated code is tied to the compiler
and less suited for dynamic structures. rfw keeps templates declarative and
uses the same patcher regardless of component complexity.

### Change-detection systems (Angular)
Angular traverses component trees on every tick, checking bindings for changes.
This works without a virtual DOM but can trigger more work than necessary and
relies on zone-based heuristics.

rfw's Selective DOM Patching combines the directness of hand-written DOM code
with the convenience of declarative templates. Together with its explicit
store-driven reactivity model, it offers a unique, minimal and solid approach
that has no direct counterpart in other frameworks.
