//go:build js && wasm

package registry

import (
	components "github.com/rfwlab/rfw/docs/examples/components"
	plugins "github.com/rfwlab/rfw/docs/examples/plugins"
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/router"
	"github.com/rfwlab/rfw/v2/state"
)

// Demo describes one runnable framework example. The docs app uses the same
// registry for route registration, demo indexes, and API coverage tracking.
type Demo struct {
	Path      string
	Title     string
	Track     string
	APIs      []string
	Source    string
	Component func() core.Component
	Guards    []router.Guard
	Children  []router.Route
}

// Routes returns router routes for all examples.
func Routes(appStore *state.Store) []router.Route {
	demos := All(appStore)
	routes := make([]router.Route, 0, len(demos))
	for _, demo := range demos {
		routes = append(routes, router.Route{
			Path:      demo.Path,
			Component: demo.Component,
			Guards:    demo.Guards,
			Children:  demo.Children,
		})
	}
	return routes
}

// All returns every demo in a stable, grouped order.
func All(appStore *state.Store) []Demo {
	return []Demo{
		{Path: "/examples/main", Title: "Main Component", Track: "components", APIs: []string{"core", "components", "composition"}, Source: "/examples/components/main_component.go", Component: func() core.Component { return components.NewMainComponent() }},
		{Path: "/examples/test", Title: "Template Features", Track: "templates", APIs: []string{"rtml", "state"}, Source: "/examples/components/test_component.go", Component: func() core.Component { return components.NewTestComponent() }},
		{Path: "/examples/dynamic", Title: "Dynamic Components", Track: "components", APIs: []string{"core.ComponentRegistry", "rt-is"}, Source: "/examples/components/dynamic_component.go", Component: func() core.Component { return components.NewDynamicComponent() }},
		{Path: "/examples/slots", Title: "Slots", Track: "templates", APIs: []string{"slots", "core.AddDependency"}, Source: "/examples/components/slot_parent_component.go", Component: func() core.Component { return components.NewSlotParentComponent(nil) }},
		{Path: "/examples/user/:name", Title: "Route Params", Track: "routing", APIs: []string{"router", "OnParams"}, Source: "/examples/components/another_component.go", Component: func() core.Component { return components.NewAnotherComponent() }},
		{Path: "/examples/params/:id", Title: "Params Component", Track: "routing", APIs: []string{"router", "SetRouteParams"}, Source: "/examples/components/params_component.go", Component: func() core.Component { return components.NewParamsComponent() }},
		{
			Path:      "/examples/event",
			Title:     "Events",
			Track:     "browser",
			APIs:      []string{"events", "composition.On"},
			Source:    "/examples/components/event_component.go",
			Component: func() core.Component { return components.NewEventComponent() },
			Children: []router.Route{
				{Path: "/examples/event/listener", Component: func() core.Component { return components.NewEventListenerComponent() }},
				{Path: "/examples/event/observer", Component: func() core.Component { return components.NewObserverComponent() }},
			},
		},
		{Path: "/examples/computed", Title: "Computed Values", Track: "reactivity", APIs: []string{"state.Map2", "state.Watcher"}, Source: "/examples/components/computed_component.go", Component: func() core.Component { return components.NewComputedComponent() }},
		{Path: "/examples/webgl", Title: "WebGL Snake", Track: "game", APIs: []string{"webgl", "game/loop", "game/scene"}, Source: "/examples/components/webgl_component.go", Component: func() core.Component { return components.NewWebGLComponent() }},
		{Path: "/examples/animations", Title: "Animations", Track: "game", APIs: []string{"animation"}, Source: "/examples/components/animation_component.go", Component: func() core.Component { return components.NewAnimationComponent() }},
		{Path: "/examples/input", Title: "Input Manager", Track: "browser", APIs: []string{"input"}, Source: "/examples/components/input_component.go", Component: func() core.Component { return components.NewInputComponent() }},
		{Path: "/examples/cinema", Title: "Cinema Builder", Track: "game", APIs: []string{"animation.CinemaBuilder"}, Source: "/examples/components/cinema_component.go", Component: func() core.Component { return components.NewCinemaComponent() }},
		{Path: "/examples/shortcut", Title: "Shortcuts", Track: "plugins", APIs: []string{"plugins/shortcut"}, Source: "/examples/components/shortcut_component.go", Component: func() core.Component { return components.NewShortcutComponent() }},
		{Path: "/examples/plugins", Title: "Plugin Hooks", Track: "plugins", APIs: []string{"core.Plugin"}, Source: "/examples/plugins/plugins_component.go", Component: func() core.Component { return plugins.NewPluginsComponent() }},
		{Path: "/examples/toast", Title: "Toast Plugin", Track: "plugins", APIs: []string{"plugins/toast"}, Source: "/examples/plugins/toast_component.go", Component: func() core.Component { return plugins.NewToastComponent() }},
		{Path: "/examples/plugin-directives", Title: "Plugin Directives", Track: "plugins", APIs: []string{"core.Plugin", "directives"}, Source: "/examples/components/plugin_directives_component.go", Component: func() core.Component { return components.NewPluginDirectivesComponent() }},
		{Path: "/examples/stores", Title: "Stores", Track: "reactivity", APIs: []string{"state.Store", "state.WithPersistence"}, Source: "/examples/components/stores_component.go", Component: func() core.Component { return components.NewStoresComponent() }},
		{
			Path:      "/examples/parent",
			Title:     "Nested Components",
			Track:     "components",
			APIs:      []string{"router.Children", "core.AddDependency"},
			Source:    "/examples/components/parent_component.go",
			Component: func() core.Component { return components.NewParentComponent() },
			Children:  []router.Route{{Path: "/examples/parent/child", Component: func() core.Component { return components.NewChildComponent() }}},
		},
		{
			Path:      "/examples/protected",
			Title:     "Route Guards",
			Track:     "routing",
			APIs:      []string{"router.Guard"},
			Source:    "/examples/components/protected_component.go",
			Component: func() core.Component { return components.NewProtectedComponent() },
			Guards: []router.Guard{func(params map[string]string) bool {
				allowed, _ := appStore.Get("allowProtected").(bool)
				return allowed
			}},
		},
		{Path: "/examples/complex/:user/:section", Title: "Complex Routing", Track: "routing", APIs: []string{"router params"}, Source: "/examples/components/complex_routing_component.go", Component: func() core.Component { return components.NewComplexRoutingComponent() }},
		{Path: "/examples/state", Title: "State Management", Track: "reactivity", APIs: []string{"state.Action", "state.Map"}, Source: "/examples/components/state_management_component.go", Component: func() core.Component { return components.NewStateManagementComponent() }},
		{Path: "/examples/state-bindings", Title: "State Bindings", Track: "reactivity", APIs: []string{"state", "rtml bindings"}, Source: "/examples/components/state_bindings_component.go", Component: func() core.Component { return components.NewStateBindingsComponent() }},
		{Path: "/examples/api", Title: "API Integration", Track: "data", APIs: []string{"http.FetchJSON"}, Source: "/examples/components/api_integration_component.go", Component: func() core.Component { return components.NewAPIIntegrationComponent() }},
		{Path: "/examples/fetchjson", Title: "Fetch JSON", Track: "data", APIs: []string{"http.FetchJSON"}, Source: "/examples/components/fetchjson_component.go", Component: func() core.Component { return components.NewFetchJSONComponent() }},
		{Path: "/examples/signal-bindings", Title: "Signal Bindings", Track: "reactivity", APIs: []string{"state.Signal"}, Source: "/examples/components/signal_bindings_component.go", Component: func() core.Component { return components.NewSignalBindingsComponent() }},
		{Path: "/examples/signals", Title: "Signals & Effects", Track: "reactivity", APIs: []string{"state.Signal", "state.Effect"}, Source: "/examples/components/signals_effects_component.go", Component: func() core.Component { return components.NewSignalsEffectsComponent() }},
		{Path: "/examples/pathfinding", Title: "Pathfinding", Track: "game", APIs: []string{"ai/pathfinding"}, Source: "/examples/components/pathfinding_component.go", Component: func() core.Component { return components.NewPathfindingComponent() }},
		{Path: "/examples/netcode", Title: "Netcode", Track: "server", APIs: []string{"netcode", "hostclient"}, Source: "/examples/components/netcode_component.go", Component: func() core.Component { return components.NewNetcodeComponent() }},
		{Path: "/examples/multiplayer", Title: "Multiplayer", Track: "server", APIs: []string{"netcode", "host", "game/draw"}, Source: "/examples/components/multiplayer_component.go", Component: func() core.Component { return components.NewMultiplayerComponent() }},
		{Path: "/examples/runtime-error", Title: "Runtime Error Overlay", Track: "errors", APIs: []string{"core error overlay"}, Source: "/examples/components/runtime_error_component.go", Component: func() core.Component { return components.NewRuntimeErrorComponent() }},
	}
}
