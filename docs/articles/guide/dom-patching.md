# Selective DOM Patching

**rfw** updates the page using *Selective DOM Patching*, a minimal diffing routine that works directly on real DOM nodes. When a component renders, the framework parses the new markup into a fragment and walks it alongside the existing subtree:

* Elements annotated with `[key {expr}]` are matched and moved if needed
* Text and attributes are updated in place
* Missing nodes are created and obsolete nodes removed

The browser’s DOM is the single source of truth—there’s no virtual tree to allocate or reconcile. The patcher simply mutates what already exists.

## Benefits of Skipping a Virtual DOM

* **Lower memory use**: no duplicate representation of the interface
* **Smaller runtime**: patcher fits in a handful of functions
* **Immediate feedback**: updates apply directly to nodes

## Comparisons

### Virtual DOM libraries (React, Vue)

Maintain a mirror tree in JavaScript and diff it to compute patches. Powerful but adds memory churn and abstraction overhead.

### Compile-time DOM generation (Svelte)

Compiles templates into imperative DOM instructions. Fast at runtime but tightly bound to compiler output and less flexible for dynamic structures. rfw keeps templates declarative and relies on the same patcher for any complexity.

### Change-detection systems (Angular)

Walk component trees on every tick, checking bindings. No virtual DOM, but often more work than necessary and relies on zone-based heuristics.

## rfw’s Approach

Selective DOM Patching combines the efficiency of direct DOM updates with declarative templates. Together with explicit, store-driven reactivity, it offers a minimal, solid model without a direct counterpart in other frameworks.
