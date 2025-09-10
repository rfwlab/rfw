# bundler plugin

## Why
The `bundler` plugin runs after a build to minify JavaScript, CSS and HTML files using the Go-based [`tdewolff/minify`](https://github.com/tdewolff/minify) library. It keeps shipped assets small without external tooling.

## When to use
Enable it whenever your project has files under `static/` that should be compressed in the final `build/` directory. The plugin walks `build/` and processes but does not bundle or resolve imports:
- `.js` files with a JavaScript minifier
- `.css` files with a CSS minifier
- `.html` files with an HTML minifier

Inline `<script>` and `<style>` blocks inside `.rtml` templates are minified automatically during rendering and are not handled by this plugin.

## How
Enable the plugin in `rfw.json`:

```json
{
  "plugins": {
    "bundler": {}
  }
}
```

Run `rfw build`. During `PostBuild` the plugin rewrites the build output with minified assets.

## Example

`build/static/app.js` before:

```js
function add ( a , b ){ return a + b ; }
```

After `rfw build`:

```js
function add(a,b){return a+b}
```

## Notes
- CSS files containing Tailwind directives (e.g. `@tailwind` or `tailwindcss`) are skipped so the `tailwind` plugin can process them.
