# rfw vs Vue, React, Svelte, An Honest Comparison

This comparison is written honestly. rfw is a newer framework with a smaller ecosystem. It has real strengths, type safety, single-language stack, fine-grained reactivity, built-in SSC, but also real gaps. Use this to decide whether rfw fits your project.

---

## Philosophy

| Framework | Core idea |
|-----------|-----------|
| **Vue** | Progressive framework. Incrementally adoptable. Approachable reactivity with proxies. |
| **React** | UI as a function of state. Declarative rendering via virtual DOM diffing. Ecosystem-first. |
| **Svelte** | Compile away the framework. Minimal runtime. Reactive assignments are just variable mutations. |
| **rfw** | Single-language (Go) full-stack. Type-safe signals. Fine-grained DOM updates without virtual DOM. Server-Side Computed by default. |

Vue and React optimize for JavaScript ecosystem breadth. Svelte optimizes for minimal bundle size. rfw optimizes for type safety and server-integrated rendering in a single language.

---

## Language

| | Vue / React / Svelte | rfw |
|---|---|---|
| **Language** | JavaScript / TypeScript | Go |
| **Type safety** | Opt-in via TypeScript; runtime type errors still possible | Mandatory; compiler enforces types at build time |
| **Runtime errors** | Common (undefined is not a function, etc.) | Rare, most caught at compile time |
| **Server language** | JS/TS (Node.js) or different language | Go, same as client |
| **Shared types** | Possible with TS isomorphic code | Built-in, same Go types on server and client |
| **Interoperability** | npm ecosystem (2M+ packages) | Go ecosystem + JS interop via `js` package |

Go's type system catches what TypeScript can't: exhaustive switch checks, nil pointer safety (with care), and struct field validation. But you lose access to npm packages on the client side. You can call JavaScript from WASM via the `js` interop layer, but it's manual.

---

## Bundle Size

| Framework | Typical client bundle | Notes |
|-----------|----------------------|-------|
| **Vue** | ~33 KB gzipped (runtime + compiler) | Smaller with production build |
| **React** | ~42 KB gzipped (react + react-dom) | Plus router, state lib |
| **Svelte** | ~2–10 KB gzipped | Compiler strips framework; per-component overhead is tiny |
| **rfw** | WASM binary (~1–5 MB compressed) | Go WASM binary includes runtime; Brotli compression helps |

rfw's WASM bundle is significantly larger than any JS framework's output. This is the primary performance trade-off. WASM initial download and parse time is higher, though execution speed can be faster after loading.

Mitigations: Brotli compression (`.wasm.br`), streaming compilation, SSC pre-renders content so users see HTML before WASM loads.

---

## Reactivity Model

| Feature | Vue | React | Svelte | rfw |
|---------|-----|-------|--------|-----|
| **Mechanism** | Proxy-based interception | Virtual DOM diffing | Compile-time tracking | Explicit signals |
| **Granularity** | Component-level re-render | Component-level re-render | DOM-level updates | DOM-level updates |
| **State declaration** | `ref()` / `reactive()` / `computed()` | `useState()` / `useReducer()` | `let` / `$:` | `*t.Int` etc (type-based detection) |
| **Mutation** | Direct (proxied) | Immutable (setState) | Direct assignment | `.Set()` method |
| **Read** | Direct (proxied) | Variable read | Variable read | `.Get()` method |
| **Derived state** | `computed()` | `useMemo()` | `$:` | Go methods / `@expr:` / `state.Map()` |
| **Side effects** | `watch()` / `watchEffect()` | `useEffect()` | `$:` reactive statements | `state.Effect()` / `OnMount()` |
| **Auto-tracking** | Yes (proxy intercepts) | No (dependency array) | Yes (compiler) | No (explicit `.Get()` required) |

Svelte and Vue are more ergonomic, you just assign to variables. React is more explicit with setter functions and dependency arrays. rfw is the most explicit: every reactive read and write is a method call on a typed signal. This reduces accidental reactivity bugs at the cost of verbosity.

---

## SSR / SSC

| Feature | Nuxt (Vue) | Next.js (React) | SvelteKit | rfw |
|---------|-----------|----------------|-----------|-----|
| **SSR mode** | Optional | Optional (SSR, SSG, ISR) | Optional | Required (SSC) |
| **SPA fallback** | Yes | Yes | Yes | No |
| **Data fetching** | `useAsync` / `useFetch` | `getServerSideProps` / RSC | `load` function | Host component handler |
| **Transport** | HTTP (per request) | HTTP (per request) | HTTP (per request) | WebSocket (persistent) |
| **Live updates** | No (requires polling/SSE) | No (requires polling/SSE) | No (requires polling/SSE) | Yes (built-in via WebSocket) |
| **Shared code** | JS/TS isomorphic | JS/TS isomorphic | JS/TS isomorphic | Go (same types, same binary) |
| **Hydration** | HTML → Vue runtime | HTML → React runtime | HTML → Svelte runtime | HTML → WASM hydration |
| **Server-side secrets** | Must be careful (bundle leak risk) | Must be careful | Must be careful | Safe, server code never ships to client |

rfw's SSC model is fundamentally different from SSR in JS frameworks. The persistent WebSocket enables real-time server-to-client updates without additional infrastructure. Server-side code is genuinely private, it never ships to the browser. The trade-off is that SSC is mandatory: there is no client-only mode.

---

## Ecosystem Maturity

Be honest: this is rfw's weakest area.

| | Vue | React | Svelte | rfw |
|---|---|---|---|---|
| **npm/Go packages** | 2M+ | 2M+ | Smaller subset | Go ecosystem + js interop |
| **Component libraries** | Vuetify, Element, Naive, etc. | MUI, Chakra, Ant Design | Smaller selection | None yet |
| **UI frameworks** | Many | Many | Fewer | Use Tailwind or custom CSS |
| **DevTools** | Vue DevTools (excellent) | React DevTools (excellent) | Svelte DevTools (basic) | Basic CLI debug mode |
| **HMR** | Vite (near-instant) | Fast Refresh (~1s) | Vite (~200ms) | `rfw dev` rebuild (seconds) |
| **IDE support** | Volar, Vetur | Built-in everywhere | Svelte for VS Code | Go LSP (excellent for Go) |
| **Community size** | Large | Very large | Growing | Small |
| **Production usage** | Widespread | Widespread | Growing | Early adopters |
| **Learning resources** | Extensive | Extensive | Good | Limited (this docs set) |

If you need rich UI component libraries or extensive third-party integrations, Vue or React are safer choices today.

---

## Developer Experience

| Feature | Vue | React | Svelte | rfw |
|---------|-----|-------|--------|-----|
| **Setup** | `npm create vue` | `create-react-app` / Vite | `npm create svelte` | `rfw init` |
| **Hot reload speed** | Fast (~50ms) | Fast (~200ms) | Fast (~100ms) | Rebuild (~1–3s) |
| **Type checking** | TypeScript (opt-in) | TypeScript (opt-in) | TypeScript (opt-in) | Go compiler (mandatory) |
| **Build time** | ~1–5s | ~1–5s | ~1–3s | ~3–10s (WASM compilation) |
| **Error messages** | Good | Good | Good | Go compiler (very precise) |
| **Debugging** | Browser DevTools | Browser DevTools | Browser DevTools | `rfw dev --debug` + Go debugger |
| **Testing** | Vitest / Jest / Cypress | Jest / RTL / Cypress | Vitest / Cypress | Go `testing` package |
| **Linting** | ESLint | ESLint | ESLint | `go vet` / `golangci-lint` |
| **Formatting** | Prettier | Prettier | Prettier | `gofmt` (built-in) |

rfw's HMR is slower than Vite-based JS frameworks because it must recompile Go to WASM on each change. The Go compiler's error messages are excellent and catch issues JS frameworks only discover at runtime.

---

## When to Choose rfw

**Choose rfw if:**

- Your team already writes Go and wants a single-language stack.
- Type safety is critical, you want compile-time guarantees, not runtime surprises.
- You need tight server-client integration with real-time updates (the WebSocket-first model fits dashboards, collaborative tools, live data).
- You want server-side secrets to never ship to the browser.
- You're building an internal tool, dashboard, or data-heavy app where the larger WASM bundle is acceptable.

**Stay with Vue/React/Svelte if:**

- You need rich UI component libraries (data tables, date pickers, chart libraries).
- Your team doesn't know Go and the learning curve isn't worth it.
- Initial page load speed is critical and you need sub-100KB bundles.
- You rely heavily on npm packages for specific functionality (maps, editors, specialized widgets).
- You need a large hiring pool, Vue/React developers are far more available.
- You want progressive enhancement or client-only SPAs.

**Consider migrating incrementally if:**

- You have a Go backend and want to share types/validation with the frontend.
- You're hitting TypeScript limitations (runtime type errors despite TS).
- You want to replace a complex SSR setup (Next.js/Nuxt) with a simpler, unified model.

---

## Feature-by-Feature Comparison

| Feature | Vue | React | Svelte | rfw |
|---------|-----|-------|--------|-----|
| Language | JS/TS | JS/TS | JS/TS | Go |
| Runtime model | Virtual DOM | Virtual DOM | No runtime (compiled) | Signals + WASM |
| Bundle size | Small | Medium | Tiny | Large (WASM) |
| Type safety | Opt-in (TS) | Opt-in (TS) | Opt-in (TS) | Mandatory (Go) |
| Reactivity | Proxy-based | State + diffing | Compiler-tracked | Explicit signals |
| DOM updates | Component-level | Component-level | DOM-level | DOM-level |
| SSR/SSC | Optional (Nuxt) | Optional (Next.js) | Optional (SvelteKit) | Required (SSC) |
| Server-client transport | HTTP | HTTP | HTTP | WebSocket |
| Live server updates | No (extra infra) | No (extra infra) | No (extra infra) | Yes (built-in) |
| Same language server/client | Yes (Node) | Yes (Node) | Yes (Node) | Yes (Go) |
| Shared types | Yes (TS isomorphic) | Yes (TS isomorphic) | Yes (TS isomorphic) | Yes (Go, no serialization) |
| Scoped CSS | Yes | Yes (CSS Modules) | Yes (built-in) | No (external) |
| Component libraries | Many | Many | Few | None yet |
| DevTools | Excellent | Excellent | Basic | Basic |
| HMR speed | Fast | Fast | Fast | Slow (recompile) |
| Testing frameworks | Many | Many | Several | Go testing |
| Ecosystem size | Large | Very large | Medium | Small |
| Production readiness | Widespread | Widespread | Growing | Early |
| Form handling | Manual / libraries | Manual / libraries | `bind:` + `enhance` | `@signal:..:w` + host |
| Routing | vue-router | react-router / Next.js | SvelteKit | router.Page/Group |
| State management | Pinia / Vuex | Redux / Zustand / Jotai | Stores | signals + state.Store |
| DI | provide/inject | Context | setContext/getContext | `*t.Inject[T]` + Container |
| i18n | vue-i18n | react-i18next | svelte-i18n | rfw i18n plugin |
| SEO | Nuxt SSR | Next.js SSR | SvelteKit SSR | SSC (built-in) |
| Mobile | Capacitive / NativeScript | React Native | Svelte Native | Not supported |

---

## Summary

rfw trades:
- **Ecosystem breadth** for type safety
- **Bundle size** for a single-language stack
- **HMR speed** for compile-time guarantees
- **UI library availability** for server-integrated rendering
- **Community size** for Go-based simplicity

If these trades align with your priorities, especially if you're a Go team building data-rich applications, rfw is worth evaluating. If you need the ecosystem, the fast HMR, or the tiny bundles, Vue/React/Svelte remain the pragmatic choice.