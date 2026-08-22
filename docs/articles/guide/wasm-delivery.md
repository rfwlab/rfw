# How the bundle reaches the browser

A Go WebAssembly bundle is large. `rfw build` produces compressed artifacts and
the loader picks one, and both halves are framework-owned: an application does
not configure compression, and there is no flag to set.

This page describes the contract, because two deployment shapes need different
things from it.

## What the build produces

A production build writes three files into `build/client`:

| File | Content coding | Who can decode it |
|------|----------------|-------------------|
| `app.wasm` | none | anything |
| `app.wasm.br` | `br` | the browser's HTTP layer, only when the server labels the response |
| `app.wasm.gz` | `gzip` | the browser's HTTP layer, or the loader itself |

A development build writes only `app.wasm` and removes stale artifacts, so a
rebuild is never shadowed by a compressed copy of the previous one.

The asymmetry between the two codings matters. `DecompressionStream` implements
gzip and deflate but not brotli, so brotli only works when the server sets
`Content-Encoding: br`. Gzip works either way, which is what makes plain static
hosting possible.

## What the build tells the client

`rfw_config.js` describes the build. Every name it defines is always defined,
including when the answer is "none":

```js
window.RFW_WASM_VERSION = "e3789c452a29c4e8";
window.RFW_WASM_ENCODINGS = ["br", "gzip"];
window.RFW_WASM_NEGOTIATED = true;
window.RFW_BUILD_MODE = "production";
```

An absent global and a disabled feature are indistinguishable to the loader, so
the build never omits one. An application that reads an undefined global of its
own invention gets `undefined`, which is falsy, which silently turns off
compression: that is the failure this contract exists to prevent.

`rfw build` also stamps the build version onto the bootstrap script tags in
`index.html`:

```html
<script src="/rfw_config.js?v=e3789c452a29c4e8"></script>
<script src="/wasm_exec.js?v=e3789c452a29c4e8"></script>
<script src="/wasm_loader.js?v=e3789c452a29c4e8"></script>
```

The configuration cannot version the tag that loads it, so the markup carries
the version instead. Without this a cached loader from an earlier release keeps
driving new bundles. Tags the build does not recognise are left untouched.

## Serving it: a live host

`host.NewMux` negotiates. A request for `/app.wasm` is answered with the best
artifact the client advertised in `Accept-Encoding`, labelled with
`Content-Encoding`, `Content-Type: application/wasm`, the compressed
`Content-Length`, and `Vary: Accept-Encoding`.

This is why the loader needs no knowledge of the page's security context.
Browsers only advertise `br` on secure origins, so an HTTPS page gets brotli and
a plain HTTP page gets gzip without anything client-side deciding it. A client
that advertises nothing, or refuses both codings with `q=0`, gets the raw
bundle, which is the only body it is guaranteed to understand.

Caching follows the URL. A request carrying `?v=<hash>` is immutable for a year;
one without it must revalidate. `rfw_config.js` always revalidates, because it
is the pointer that names the current version.

## Serving it: static hosting

A CDN or object store will not negotiate. `rfw.json` with `"build": {"type":
"static"}` sets `RFW_WASM_NEGOTIATED = false`, and the loader then requests a
named artifact instead.

If the host sets `Content-Encoding` for `.br` and `.gz` files, everything works
as it does with a live host. If it does not, the loader detects the unlabelled
response and decodes `app.wasm.gz` itself through `DecompressionStream`. Brotli
cannot be rescued this way, so a static host that does not label its files falls
back to gzip. Configure the header if you can; the bundle is roughly a fifth
smaller under brotli.

## What the loader will not do

In a production build the raw bundle is not a candidate. If every compressed
path fails, the loader raises a visible error naming what it tried rather than
quietly downloading several megabytes of uncompressed wasm. A development build
does keep the raw bundle as a fallback, since that is the only artifact it has.

The progress bar is honest about what it knows. An encoded response reports the
compressed length while the reader yields decoded bytes, so a percentage
computed from the two is fiction; the loader shows an indeterminate state
instead. It shows a real percentage only for an unencoded body, where the two
numbers measure the same thing.

## Checking a deployment

From the built client directory:

```bash
curl -sI -H 'Accept-Encoding: br, gzip' http://localhost:8080/app.wasm
```

Expect `Content-Encoding: br`, `Content-Type: application/wasm`, a
`Content-Length` matching `app.wasm.br` on disk, and `Vary: Accept-Encoding`.
Repeat with `-H 'Accept-Encoding: gzip'` and expect gzip. A response with no
`Content-Encoding` for a compression-capable client means the artifacts are
missing or the server in front of rfw stripped them.
