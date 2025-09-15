# Assets Plugin

The **Assets plugin** copies static files into the build output so they can be served with your application.

## Features

* Copies images, fonts, and other static resources into your build.
* Works with both `rfw build` and `rfw dev`.
* Rebuilds automatically when assets change during development.

## Configuration

Enable the plugin in `rfw.json`:

```json
{
  "plugins": {
    "assets": {
      "dir": "assets",
      "dest": "dist"
    }
  }
}
```

Options:

* **dir** – source directory for static files (default: `assets`)
* **dest** – destination directory inside the build output (default: `dist`)

## How It Works

* On **`rfw build`**, the plugin copies all files from `dir` into `dest`.
* On **`rfw dev`**, changes in `dir` trigger a rebuild so assets stay in sync.

## When to Use

Use the Assets plugin when you want to bundle static resources with your app, without extra configuration or external tooling.
