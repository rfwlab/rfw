# router

Client-side router with lazy loaded components and guards.

- `RegisterRoute(Route)` adds a route definition.
- `Navigate(path)` programmatically changes the URL.
- Guards: `Guard` functions run before navigation and can cancel by returning `false`.
