# RTML for Visual Studio Code

Syntax highlighting and snippets for [rfw](https://github.com/rfwlab/rfw) `.rtml` component templates.

## Features

- Full HTML highlighting (the grammar extends `text.html.basic`).
- Highlighting for every rtml directive, matching the actual rfw parser:
  - Control flow: `@if:` / `@else-if:` / `@else` / `@endif`, `@for:item[,key] in expr` / `@endfor`, `@slot[:name[.modifier]]` / `@endslot`
  - Bindings: `@store:module.store.key[:w]`, `@rawstore:module.store.key`, `@signal:name[:w]`, `@prop:name`, `@rawprop:name`
  - Events: `@on:event[.modifier]:handler` and the `@event:handler` shorthand (e.g. `@click:save`, `@on:keydown.enter:submit`)
  - Composition: `@include:Component` and `@include:Component:{key:"value"}`
  - Expressions: `{{ ... }}` interpolation and `@expr:...`, including the word operators `and`, `or`, `not`, `is`, `is not`, `then`, `else` and `store:` / `signal:` / `prop:` references
  - Host components: `{h:name}` placeholders and `@h:command`
  - Plugins: `@plugin:module.command` and `[plugin:name]` markers
  - Constructor markers inside start tags: `[ref]` and `[key expr]`
- Snippets for the common building blocks: `root`, `@for`, `@if`, `@ifelse`, `@on`, `@store`, `@include`, `@slot` and more.
- Bracket matching, autoclosing pairs, HTML comment toggling and folding for `@if` / `@for` / `@slot` blocks.

## Install from source

The extension lives in `editors/vscode/` of the rfw repository and has no build step.

### With vsce (recommended)

```sh
cd editors/vscode
npx --yes @vscode/vsce package
code --install-extension rtml-0.1.0.vsix
```

### Manual copy

Copy the directory into your local extensions folder:

```sh
mkdir -p ~/.vscode/extensions
cp -r editors/vscode ~/.vscode/extensions/rfwlab.rtml-0.1.0
```

Then restart VS Code (or run the "Developer: Reload Window" command). Files ending in `.rtml` will pick up the RTML language automatically.

## License

AGPL-3.0-only, same as the rfw repository. See [LICENSE](./LICENSE).
