# Server‑Side Computed (SSC)

**SSC** runs most application logic on the **server**, while the browser loads a lightweight **Wasm** bundle to hydrate server‑rendered HTML. Server and client keep state synchronized over a persistent **WebSocket**. In rfw, server‑originated bindings and commands are prefixed with `h:`.

---

## What is SSC?

In a standard client‑side app, components render and update entirely in the browser. With SSC, the **host** process (a Go server) renders HTML and executes privileged logic. The browser receives fully rendered markup, then the Wasm bundle **hydrates** it—attaching event handlers and reactive bindings—so the page becomes interactive. Updates that involve server state travel over the WebSocket.

**Key idea:** keep heavy / sensitive work on the server; keep the client tiny and reactive.

---

## Why SSC?

Compared to a pure client‑side SPA, SSC typically offers:

* **Faster time‑to‑content**: users see meaningful HTML immediately, before Wasm finishes loading.
* **Unified Go model**: write both server and client logic in Go; share types and business rules.
* **Better SEO**: crawlers can index the server‑rendered HTML.
* **Tighter security**: secrets and privileged actions remain on the server.

> Note: exact performance depends on your app and infrastructure.

---

## Trade‑offs

* **More moving parts**: you build **two** artifacts: a client Wasm bundle and a **host** binary.
* **Server resources**: rendering and synchronizing state consumes CPU / memory on the server; plan caching and capacity.
* **Environment constraints**: browser‑only code must run on the client side; server code cannot use browser globals.
* **Hydration care**: server HTML and client expectations must match to avoid hydration mismatches.

Where these trade‑offs are acceptable—dashboards, content sites, apps with strict latency / SEO needs—SSC shines.

---

## Enabling SSC

Set the build type in `rfw.json` (created by `rfw init` with `ssc` by default):

```json
{
  "build": { "type": "ssc" }
}
```

Running `rfw build` produces:

* `build/client/` – the Wasm bundle and client assets
* `build/host/` – the **host** server binary that serves the client and synchronizes state

With `ssc` active, variables and commands prefixed with `h:` are kept in sync over a WebSocket.

---

## Minimal Tutorial

### 1) HTML component opts into a host component

```go
//go:build js && wasm
package components

import (
    _ "embed"
    core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/greeting_component.rtml
var greetingTpl []byte

func NewGreetingComponent() *core.HTMLComponent {
    c := core.NewComponent("GreetingComponent", greetingTpl, map[string]any{
        "clientMsg": "hello from wasm",
    })
    c.AddHostComponent("GreetingHost") // opt into server logic
    return c
}
```

### 2) Host registers the matching component and serves

```go
package main

import "github.com/rfwlab/rfw/v1/host"

func main() {
    host.Register(host.NewHostComponent("GreetingHost", func(_ map[string]any) any {
        return map[string]any{"hostMsg": "hello from server"}
    }))
    host.ListenAndServe(":8080", "client") // serves bundle and /ws
}
```

### 3) Template mixing client and host values

```rtml
<root>
  <p>Client: {clientMsg}</p>
  <p>Host: {h:hostMsg}</p>
  <button @on:click:h:updateTime>refresh</button>
</root>
```

* `{clientMsg}` ships with the Wasm component.
* `{h:hostMsg}` is pushed by the host over the WebSocket.
* `@on:click:h:updateTime` invokes a **host** command.

---

## Development Workflow

* `rfw dev` rebuilds on changes, serves assets, and—when a `host/` directory is present—builds and runs the host binary from `build/host/host` so you can iterate locally.
* Use `--debug` to enable profiling endpoints (served by both the dev server and any host binary) and load the DevTools overlay.

> Ports, logs, and profiling flags follow the standard dev server behavior. See **Dev Server** docs.

---

## Hydration: how the page becomes interactive

1. The host responds with **fully rendered HTML**.
2. The browser downloads `app.wasm` and initializes rfw.
3. rfw **hydrates** the server markup: attaches event handlers and reactive bindings.
4. A WebSocket connects; `h:` values and commands synchronize with the host.

**Avoid mismatches:** ensure the data that renders on the server is consistent with what the client expects on first paint (e.g., avoid random values and time‑zone‑dependent formatting during initial render). If divergence is unavoidable, render those parts client‑only after mount.


---

## Data Flow & API Surface

* **Bindings**: `{h:key}` reads server‑side values provided by the host component.
* **Commands**: `@on:click:h:cmd` (and similar) invoke host handlers.
* **Transport**: a persistent WebSocket at `/ws` carries `h:` updates and commands.

> Exact message shapes are internal; only the `h:` prefix contract is relevant when authoring components.

---

## Code Structure (sharing types & logic)

Keep shared types and validation in regular Go packages and import them from both sides:

* **Shared**: DTOs, validators, domain logic
* **Host only**: database access, secrets, privileged integrations
* **Client only**: DOM‑dependent code, animations, browser APIs

Run server‑only code in the **host**; run browser‑only code in the Wasm bundle. Avoid assuming browser globals on the host.

> Per‑request store isolation / instancing details are **not specified in documentation**. Prefer stateless host handlers and pass all request‑specific data explicitly.

---

## SSC vs. SSG

**SSG** (static site generation) pre‑renders pages at build time and serves static HTML. It’s ideal for content that doesn’t change per user and for ultra‑cheap hosting.

* rfw’s built‑in static plugins (e.g., **assets**, **docs**) help ship content, but a dedicated SSG pipeline is **not specified in documentation**.
* Choose SSC when you need **per‑request data**, authenticated content, or server‑pushed updates.

---

## Caching & Performance

* Cache host‑generated responses or fragments when possible.
* Coalesce frequent updates server‑side and send compact diffs through `h:` bindings.
* Profile with `--debug` (`/debug/vars`, `/debug/pprof/`) during development.

Concrete caching hooks / helpers are **not specified in documentation**; implement with standard Go middleware or reverse proxies.

---

## Debugging

* Use the DevTools overlay (enable with `--debug`) to observe component trees and network helpers.
* Log host activity in your handlers; surface key values through `{h:*}` to verify synchronization.

WebSocket diagnostics beyond standard logs are **not specified in documentation**.

---

## Deployment

* Deploy the **host** binary behind your reverse proxy; serve the **client** bundle from `build/client/`.
* Ensure the WebSocket endpoint (`/ws`) is reachable and upgraded by your proxy.
* Configure TLS / CORS as needed at the proxy or host level.

Detailed deployment blueprints are **not specified in documentation**.

---

## FAQ

**How do I opt‑in per component?**  Call `AddHostComponent("Name")` on the HTML component and register a host component with the same name on the server.

**Can I mix SSC and client‑only components?**  Yes—only components that declare a host partner participate in SSC; others remain client‑only.

**Where do `h:` values come from?**  From the host component’s handler; they are pushed over the WebSocket and exposed to templates as `{h:key}`.

**Can the host trigger a client action?**  Yes—by updating `h:` values or responding to `h:` commands; ad‑hoc client‑initiated events still live in Wasm.

**Is streaming / partial hydration supported?**  **Not specified in documentation.**

---

## Full Example (recap)

**Client component**

```go
//go:build js && wasm
package components

import (
    _ "embed"
    core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/greeting_component.rtml
var greetingTpl []byte

func NewGreetingComponent() *core.HTMLComponent {
    c := core.NewComponent("GreetingComponent", greetingTpl, map[string]any{
        "clientMsg": "hello from wasm",
    })
    c.AddHostComponent("GreetingHost")
    return c
}
```

**Host**

```go
package main

import "github.com/rfwlab/rfw/v1/host"

func main() {
    host.Register(host.NewHostComponent("GreetingHost", func(_ map[string]any) any {
        return map[string]any{"hostMsg": "hello from server"}
    }))
    host.ListenAndServe(":8080", "client")
}
```

**Template**

```rtml
<root>
  <p>Client: {clientMsg}</p>
  <p>Host: {h:hostMsg}</p>
  <button @on:click:h:updateTime>refresh</button>
</root>
```

---

## See also

* [Manifest](../manifest) – enable `ssc`
* [Host](../api/host) & [Host Client](../api/hostclient)
* [Dev Server](../devtools) – debug overlay and profiling
