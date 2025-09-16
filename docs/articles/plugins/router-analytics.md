# Router Analytics Plugin

The **Router Analytics** plugin tracks page-to-page transitions and predicts likely next routes. It can also emit **prefetch hints** for the most probable next pages, so your app can warm up data/assets.

> This is a **built‑in plugin**. It doesn’t change your routing— it observes navigations and learns probabilities.

## Features

* Records transitions between normalized routes (e.g. `/posts` → `/settings`).
* Exposes **transition probabilities** and **top‑N predictions**.
* Optional **prefetch hints** over a netcode channel (browser / Wasm builds).
* Pluggable route normalization.

## Quick Start

Register the plugin at boot:

```go
import (
  core "github.com/rfwlab/rfw/v1/core"
  routeranalytics "github.com/rfwlab/rfw/v1/plugins/routeranalytics"
)

func main() {
  core.RegisterPlugin(routeranalytics.New(routeranalytics.Options{
    // optional: customize behavior
    // PrefetchLimit: 3,                   // number of next routes to hint
    // PrefetchThreshold: 0.25,            // drop hints below this probability
    // Channel: "RouterPrefetch",          // netcode channel name
    // Normalize: routeranalytics.NormalizePath, // default normalization
  }))
}
```

The plugin automatically hooks into the app router and learns from each navigation.

## Example (predict next routes)

```go
// Get ordered probabilities for what users do next from "/posts"
probs := routeranalytics.TransitionProbabilities("/posts")
for _, p := range probs {
  // p.To, p.Count, p.Probability
}

// Ask for the top 2 most likely next routes
top2 := routeranalytics.MostLikelyNext("/posts", 2)
```

If `PrefetchLimit` > 0, each navigation will enqueue prefetch hints for predicted routes with probability ≥ `PrefetchThreshold` (browser/Wasm builds only).

## How It Works

* **Normalization** (default): trims whitespace, removes `?query` and `#hash`, and ensures a **leading slash**.
* **Learning**: each navigation updates counters for `from → to` and totals per `from`.
* **Probabilities**: computed as `count(from→to) / total(from)` and returned **sorted** by probability (then alphabetically by `to`).
* **Prefetch**: in Wasm builds, predicted routes are sent once over a netcode client on the configured channel.

## API Reference

### Types

```go
type Options struct {
  Normalize         func(string) string // normalize a path before tracking
  PrefetchLimit     int                 // number of routes to prefetch (<=0 disables)
  PrefetchThreshold float64            // drop hints below this probability (<=0 → 0.2)
  Channel           string             // netcode channel (default "RouterPrefetch")
}

type TransitionProbability struct {
  From        string
  To          string
  Count       int
  Probability float64
}
```

### Constructors & Plugin hooks

```go
func New(opts Options) *Plugin
func (p *Plugin) Name() string // "routeranalytics"
func (p *Plugin) Build(json.RawMessage) error // no-op
func (p *Plugin) Install(a *core.App)         // attaches to router
```

### Instance methods

```go
func (p *Plugin) TransitionProbabilities(from string) []TransitionProbability
func (p *Plugin) MostLikelyNext(from string, limit int) []TransitionProbability
func (p *Plugin) Reset()
```

### Package helpers (use the latest installed instance)

```go
func TransitionProbabilities(from string) []TransitionProbability
func MostLikelyNext(from string, limit int) []TransitionProbability
func Reset()
```

### Utilities

```go
// Default normalization: trim, strip query/hash, ensure leading slash.
func NormalizePath(path string) string
```

## Defaults

* `Normalize`: `NormalizePath`
* `PrefetchLimit`: `3` (set `< 0` to disable, `0` → `3`)
* `PrefetchThreshold`: `0.2` if `<= 0`
* `Channel`: `"RouterPrefetch"`
