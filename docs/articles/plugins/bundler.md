# Bundler Plugin

The **Bundler plugin** minifies JavaScript, CSS, and HTML files after a build, keeping shipped assets small without external tools.

## Features

* Runs on `rfw build` during the **PostBuild** step.
* Minifies:

  * `.js` files with a JavaScript minifier.
  * `.css` files with a CSS minifier.
  * `.html` files with an HTML minifier.
* Inline `<script>` and `<style>` inside `.rtml` are already minified at render time (not handled by this plugin).
* Skips Tailwind CSS files so the `tailwind` plugin can process them.

## Setup

Enable the plugin in `rfw.json`:

```json
{
  "plugins": {
    "bundler": {}
  }
}
```

## Usage

Run a build:

```bash
rfw build
```

The plugin rewrites the `build/` output with minified assets.

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

* Skipped automatically in `rfw dev --debug` to keep output readable.
* Use `rfw build` for production-ready minified files.
* Does **not** bundle modules or resolve imports; it only minifies existing files.
