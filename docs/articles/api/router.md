# router

Client-side router with lazy loaded components and guards.

| Function | Description |
| --- | --- |
| `RegisterRoute(Route)` | Adds a route definition. |
| `Navigate(path)` | Programmatically changes the URL. |
| `CanNavigate(path) bool` | Reports whether a path matches a registered route. |
| `InitRouter()` | Starts the router and listens for navigation events. |
| `ExposeNavigate()` | Exposes navigation to JavaScript as `goNavigate` and auto-routes internal links. |
| `NotFoundComponent` / `NotFoundCallback` | Handle unmatched routes. |
| `Reset()` | Clears registered routes and the current component. |
| `Route.Children []Route` | Nests routes under a parent. |
| `Guard` | Runs before navigation and can cancel by returning `false`. |

