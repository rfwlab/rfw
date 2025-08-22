# Server Side Computed

The project manifest (`rfw.json`) can declare the build type. Setting it to `ssc` enables Server Side Computed builds.

```json
{
  "build": {
    "type": "ssc"
  }
}
```

When `ssc` is active, `rfw build` compiles the Wasm bundle and also builds the Go sources in the `host` directory into a server binary. The server keeps variables and commands prefixed with `h:` synchronized with the client through a persistent WebSocket connection.

## HTML and Host Components

HTML components can opt into server side logic by attaching a host component. The HTML part still compiles to Wasm, while the host component runs inside the Go server.

```go
//go:build js && wasm

package components

import (
    _ "embed"
    "github.com/rfwlab/rfw/v1/core"
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

The server registers the corresponding host component and exposes it over a WebSocket:

```go
package main

import "github.com/rfwlab/rfw/v1/host"

func main() {
    host.Register(host.NewHostComponent("GreetingHost", func(_ map[string]any) any {
        return map[string]any{"hostMsg": "hello from server"}
    }))
    host.ListenAndServe(":8090")
}
```

The RTML template can read variables coming from both the HTML component and the host. Values provided by the host are prefixed with `h:`.

```rtml
<root>
    <p>Client: {clientMsg}</p>
    <p>Host: {h:hostMsg}</p>
    <button @click:h:updateTime>refresh</button>
</root>
```

`clientMsg` ships with the Wasm component, while `h:hostMsg` is pushed over the WebSocket and updates whenever the host responds. Commands starting with `h:` invoke functions on the host component.

## WebSocket synchronisation

During runtime the Wasm code opens a persistent WebSocket to the host process. Bindings and commands prefixed with `h:` travel through this channel, allowing the server to push updates or execute privileged operations.

## Wasm and hydration

In SSC builds the Wasm bundle's job is primarily to hydrate the HTML produced by the host. The bundle is smaller because most application logic executes on the server. On first load the host serves fully rendered markup; the browser downloads the Wasm and hydrates the DOM, wiring event handlers and reactive bindings. The same flow runs during local development when the host binary serves pages on your machine.

This separation keeps sensitive logic on the server while the client only runs lightweight Wasm for reactivity.
