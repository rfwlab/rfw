//go:build js && wasm

// Package router implements a simple client-side router for WebAssembly
// applications built with rfw.
package router

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	events "github.com/rfwlab/rfw/v2/events"
	js "github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/state"
	"github.com/rfwlab/rfw/v2/types"
)

// Guard is a function that determines whether navigation to a route is
// permitted based on the provided parameters.
type Guard func(map[string]string) bool

// Route describes a routing rule that maps a path to a component and optional
// guards or child routes.
//
// Component accepts three forms:
//   - A *types.View instance: reused every navigation (singleton).
//   - A func() *types.View: called each navigation to create a fresh instance.
//   - A func() core.Component: called each navigation (legacy).
type Route struct {
	Path      string
	Name      string
	Component any
	Guards    []Guard
	Children  []Route
	Loader    Loader
	Redirect  string
	Meta      map[string]any
}

// Singleton wraps a pre-created View into a Route.Component value.
// Every navigation returns the same instance, no re-creation.
func Singleton(v *types.View) any {
	return v
}

type route struct {
	pattern    string
	fullPath   string
	name       string
	regex      *regexp.Regexp
	paramNames []string
	matchNames []string
	component  core.Component
	loader     func() core.Component
	singleton  bool
	children   []route
	guards     []Guard
	dataLoader Loader
	redirect   string
	meta       map[string]any
}

// RegisteredRoute describes a registered route in a navigable tree form.
type RegisteredRoute struct {
	// Template is the path exactly as registered (can be relative for nested
	// routes).
	Template string `json:"template"`
	// Path is the fully qualified route path derived from the registration
	// hierarchy.
	Path string `json:"path"`
	// Name is the optional identifier used by URL.
	Name string `json:"name,omitempty"`
	// Params lists the dynamic parameters extracted from the template.
	Params []string `json:"params"`
	// Children contains nested routes.
	Children []RegisteredRoute `json:"children"`
	// Meta contains user-defined route metadata.
	Meta map[string]any `json:"meta,omitempty"`
}

var (
	routes             []route
	currentComponent   core.Component
	exposeNavigateOnce sync.Once
	activePathSig      = state.NewSignal("/")
	navItems           []NavItem
	scrollEnabled      = true
	scrollPositions    = map[string][2]int{}
)

// NotFoundComponent, if set, is rendered when no route matches the path.
// Accepts the same forms as Route.Component.
var NotFoundComponent any

// NotFoundCallback, if set, is invoked when navigation targets an
// unregistered route. It receives the requested path.
var NotFoundCallback func(string)

// Reset clears the router's registered routes and current component.
// It is primarily intended for use in tests to ensure a clean state.
func Reset() {
	routes = nil
	currentComponent = nil
	NotFoundComponent = nil
	NotFoundCallback = nil
	activePathSig.Set("/")
	resetNavigation()
	scrollPositions = map[string][2]int{}
}

// RegisterRoute adds a new Route to the router's configuration.
func RegisterRoute(r Route) {
	routes = append(routes, buildRoute(r))
}

func buildRoute(r Route) route {
	return buildRouteAt(r, "")
}

func buildRouteAt(r Route, parent string) route {
	fullPath := resolveRoutePath(parent, r.Path)
	segments := strings.Split(strings.Trim(fullPath, "/"), "/")
	regexParts := make([]string, len(segments))
	matchNames := []string{}

	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			name := strings.TrimPrefix(segment, ":")
			matchNames = append(matchNames, name)
			regexParts[i] = "([^/]+)"
		} else {
			regexParts[i] = regexp.QuoteMeta(segment)
		}
	}

	paramNames := []string{}
	for _, segment := range strings.Split(strings.Trim(r.Path, "/"), "/") {
		if strings.HasPrefix(segment, ":") {
			paramNames = append(paramNames, strings.TrimPrefix(segment, ":"))
		}
	}

	pathRegex := strings.Join(regexParts, "/")
	suffix := "/?$"
	if len(r.Children) > 0 {
		suffix = "(?:/|$)"
	}
	if pathRegex == "" {
		if len(r.Children) > 0 {
			suffix = ""
		} else {
			suffix = "$"
		}
	}
	pattern := "^/" + pathRegex + suffix
	var loader func() core.Component
	var singleton bool
	switch c := r.Component.(type) {
	case *types.View:
		comp := c
		loader = func() core.Component { return comp }
		singleton = true
	case func() *types.View:
		loader = func() core.Component { return c() }
	case func() core.Component:
		loader = c
	case core.Component:
		comp := c
		loader = func() core.Component { return comp }
		singleton = true
	}
	rt := route{
		pattern:    r.Path,
		fullPath:   fullPath,
		name:       r.Name,
		regex:      regexp.MustCompile(pattern),
		paramNames: paramNames,
		matchNames: matchNames,
		loader:     loader,
		singleton:  singleton,
		guards:     r.Guards,
		dataLoader: r.Loader,
		redirect:   r.Redirect,
		meta:       cloneMeta(r.Meta),
	}

	for _, child := range r.Children {
		rt.children = append(rt.children, buildRouteAt(child, fullPath))
	}

	return rt
}

// RegisteredRoutes returns the registered routes including nested children and
// resolved full paths. The data can be used for tooling and diagnostics.
func RegisteredRoutes() []RegisteredRoute {
	out := make([]RegisteredRoute, 0, len(routes))
	for i := range routes {
		out = append(out, snapshotRoute(&routes[i], ""))
	}
	return out
}

func snapshotRoute(r *route, parent string) RegisteredRoute {
	params := make([]string, len(r.paramNames))
	copy(params, r.paramNames)
	full := resolveRoutePath(parent, r.pattern)
	children := make([]RegisteredRoute, len(r.children))
	for i := range r.children {
		children[i] = snapshotRoute(&r.children[i], full)
	}
	return RegisteredRoute{
		Template: r.pattern,
		Path:     full,
		Name:     r.name,
		Params:   params,
		Children: children,
		Meta:     cloneMeta(r.meta),
	}
}

func resolveRoutePath(parent, path string) string {
	if path == "" {
		if parent == "" {
			return "/"
		}
		return parent
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	trimmed := strings.TrimPrefix(path, "/")
	if parent == "" || parent == "/" {
		return "/" + trimmed
	}
	if strings.HasSuffix(parent, "/") {
		return parent + trimmed
	}
	return parent + "/" + trimmed
}

type routeParamReceiver interface {
	SetRouteParams(map[string]string)
}

func matchRoute(routes []route, path string) (*route, []Guard, map[string]string) {
	for i := range routes {
		r := &routes[i]
		matches := r.regex.FindStringSubmatch(path)
		if matches == nil {
			if child, guards, params := matchRoute(r.children, path); child != nil {
				return child, append(r.guards, guards...), params
			}
			continue
		}
		params := map[string]string{}
		for i, name := range r.matchNames {
			if i+1 < len(matches) {
				params[name] = matches[i+1]
			}
		}
		if child, guards, childParams := matchRoute(r.children, path); child != nil {
			return child, append(r.guards, guards...), childParams
		}
		matchedPath := strings.TrimSuffix(matches[0], "/")
		if (r.loader != nil || r.redirect != "") && matchedPath == strings.TrimSuffix(path, "/") {
			return r, r.guards, params
		}
	}
	return nil, nil, nil
}

type historyMode uint8

const (
	historyNone historyMode = iota
	historyPush
	historyReplace
)

// Navigate renders the component associated with the specified path if all
// route guards allow it. The provided path may include a query string which
// will be parsed and passed to the component via SetRouteParams.
func Navigate(fullPath string) {
	navigate(fullPath, historyPush)
}

// Replace navigates without adding a new browser history entry.
func Replace(fullPath string) {
	navigate(fullPath, historyReplace)
}

func navigate(fullPath string, history historyMode) {
	var navigationErr error
	core.TryNavigate(fullPath, func() {
		navigationErr = navigateImpl(context.Background(), fullPath, history)
	})
	if navigationErr != nil && navigationIsContextError(navigationErr) {
		return
	}
}

// NavigateContext loads and commits a route or returns a navigation error.
func NavigateContext(parent context.Context, fullPath string) error {
	var navigationErr error
	core.TryNavigate(fullPath, func() {
		navigationErr = navigateImpl(parent, fullPath, historyPush)
	})
	return navigationErr
}

func navigateImpl(parent context.Context, fullPath string, history historyMode) error {
	ctx, navigation := beginNavigation(parent)
	if err := ctx.Err(); err != nil {
		return err
	}
	path := fullPath
	query := ""
	if idx := strings.Index(fullPath, "?"); idx != -1 {
		path = fullPath[:idx]
		query = fullPath[idx+1:]
	}

	r, guards, params := matchRoute(routes, path)
	if r == nil {
		if NotFoundCallback != nil {
			NotFoundCallback(fullPath)
		} else if NotFoundComponent != nil {
			if currentComponent != nil {
				saveScroll()
				core.Log().Debug("Unmounting current component: %s", currentComponent.GetName())
				core.TriggerUnmount(currentComponent)
				currentComponent.Unmount()
			}
			var c core.Component
			switch nc := NotFoundComponent.(type) {
			case *types.View:
				c = nc
			case func() *types.View:
				c = nc()
			case func() core.Component:
				c = nc()
			}
			if c != nil {
				currentComponent = c
				if liveOutlet != nil {
					liveOutlet.renderChild(c)
				} else {
					dom.UpdateDOM(c.GetID(), core.TryRender(c))
				}
				core.TryMount(c)
				core.TriggerMount(c)
				core.TriggerRouter(fullPath)
				activePathSig.Set(fullPath)
				updateHistory(history, fullPath)
				restoreScroll(fullPath, history, nil)
				commitRouteState(nil, nil)
				return nil
			}
		}
		failNavigation(ErrRouteNotFound)
		return ErrRouteNotFound
	}

	if params == nil {
		params = map[string]string{}
	}
	queryParams, queryValues := routeQuery(query)
	for key, value := range queryParams {
		params[key] = value
	}
	for _, g := range guards {
		if !g(params) {
			if currentComponent == nil && path != "/" {
				navigate("/", history)
			}
			return ErrNavigationBlocked
		}
	}

	if r.redirect != "" {
		destination, err := redirectPath(r.redirect, params)
		if err != nil {
			failNavigation(err)
			return err
		}
		redirectCtx, err := nextRedirectContext(parent)
		if err != nil {
			failNavigation(err)
			return err
		}
		return navigateImpl(redirectCtx, destination, historyReplace)
	}

	var data any
	if r.dataLoader != nil {
		state.Batch(func() {
			navigationError.Set(nil)
			navigationStatus.Set(NavigationLoading)
		})
		loaded, err := r.dataLoader(ctx, LoadContext{
			Path:   path,
			Params: cloneStringMap(params),
			Query:  queryValues,
		})
		if err != nil {
			if navigationIsCurrent(navigation) {
				failNavigation(err)
			}
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if !navigationIsCurrent(navigation) {
			return context.Canceled
		}
		data = loaded
	}

	if r.loader != nil {
		if r.singleton && r.component != nil {
			// Re-use existing instance for singleton routes.
		} else {
			r.component = r.loader()
		}
	}
	if r.component == nil {
		failNavigation(ErrRouteComponent)
		return ErrRouteComponent
	}
	if receiver, ok := r.component.(routeParamReceiver); ok {
		receiver.SetRouteParams(params)
	}
	if receiver, ok := r.component.(RouteDataReceiver); ok {
		receiver.SetRouteData(data)
	}
	saveScroll()
	if currentComponent != nil {
		core.Log().Debug("Unmounting current component: %s", currentComponent.GetName())
		core.TriggerUnmount(currentComponent)
		currentComponent.Unmount()
	}
	currentComponent = r.component
	if liveOutlet != nil {
		liveOutlet.renderChild(r.component)
	} else {
		dom.UpdateDOM(r.component.GetID(), r.component.Render())
	}
	r.component.Mount()
	core.TriggerMount(r.component)
	r.component.OnParams(params)
	core.TriggerRouter(fullPath)
	activePathSig.Set(fullPath)
	updateHistory(history, fullPath)
	restoreScroll(fullPath, history, r.meta)
	commitRouteState(data, r.meta)
	return nil
}

func navigationIsContextError(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}

func updateHistory(mode historyMode, path string) {
	switch mode {
	case historyPush:
		js.History().Call("pushState", nil, "", path)
	case historyReplace:
		js.History().Call("replaceState", nil, "", path)
	}
}

// SetScrollRestoration enables or disables router-managed scroll positions.
func SetScrollRestoration(enabled bool) {
	scrollEnabled = enabled
}

func saveScroll() {
	if !scrollEnabled || currentComponent == nil {
		return
	}
	path := activePathSig.Get()
	scrollPositions[path] = [2]int{
		js.Window().Get("scrollX").Int(),
		js.Window().Get("scrollY").Int(),
	}
}

func restoreScroll(path string, history historyMode, meta map[string]any) {
	if !scrollEnabled {
		return
	}
	if preserve, _ := meta["preserveScroll"].(bool); preserve {
		return
	}
	position := [2]int{}
	if history == historyNone {
		position = scrollPositions[path]
	}
	js.Window().Call("scrollTo", position[0], position[1])
}

// CanNavigate reports whether the specified path matches a registered route.
func CanNavigate(fullPath string) bool {
	path := fullPath
	if idx := strings.Index(fullPath, "?"); idx != -1 {
		path = fullPath[:idx]
	}
	r, _, _ := matchRoute(routes, path)
	return r != nil
}

// ExposeNavigate makes the Navigate function accessible from JavaScript and
// automatically routes internal anchor clicks.
func ExposeNavigate() {
	exposeNavigateOnce.Do(func() {
		js.ExposeFunc("goNavigate", func(this js.Value, args []js.Value) any {
			path := args[0].String()
			Navigate(path)
			return nil
		})

		events.On("click", js.Document(), func(evt js.Value) {
			link := evt.Get("target").Call("closest", "a[href]")
			if !link.Truthy() {
				return
			}
			if t := link.Get("target").String(); t != "" && t != "_self" {
				return
			}
			if link.Get("origin").String() != js.Location().Get("origin").String() {
				return
			}
			path := link.Get("pathname").String() + link.Get("search").String()
			if CanNavigate(path) {
				evt.Call("preventDefault")
				Navigate(path)
			}
		})
	})
}

// Page registers a route with path, component and optional guards.
func Page(path string, component any, guards ...Guard) {
	RegisterRoute(Route{
		Path:      path,
		Component: component,
		Guards:    guards,
	})
}

// Group creates nested routes under a common path prefix
// and registers them. Returns the parent Route for chaining.
func Group(prefix string, fn func(*GroupBuilder)) {
	b := &GroupBuilder{prefix: prefix}
	fn(b)
	RegisterRoute(Route{
		Path:     prefix,
		Children: b.children,
	})
}

// GroupBuilder collects child routes within a Group callback.
type GroupBuilder struct {
	prefix   string
	children []Route
}

// Page adds a child route within a Group.
func (g *GroupBuilder) Page(path string, component any, guards ...Guard) {
	g.children = append(g.children, Route{
		Path:      path,
		Component: component,
		Guards:    guards,
	})
}

// InitRouter initializes the router and begins listening for navigation
// events.
func InitRouter() {
	ExposeNavigate()

	// The popstate listener lives for the whole app lifetime; the stop
	// function is intentionally discarded.
	ch, _ := events.Listen("popstate", js.Window())
	go func() {
		for range ch {
			path := js.Location().Get("pathname").String() + js.Location().Get("search").String()
			navigate(path, historyNone)
		}
	}()

	currentPath := js.Location().Get("pathname").String() + js.Location().Get("search").String()
	navigate(currentPath, historyNone)
}

// NavItem describes a navigation entry with arbitrary metadata.
type NavItem struct {
	Path  string         `json:"path"`
	Label string         `json:"label"`
	Meta  map[string]any `json:"meta"`
}

// SetNavItems registers the navigation items to be consumed by templates.
func SetNavItems(items []NavItem) {
	navItems = items
}

// NavItems returns the previously registered navigation items.
func NavItems() []NavItem {
	return navItems
}

// NavItemsMap returns navigation items as []any of map[string]any,
// ready for template consumption in @for: loops.
func NavItemsMap() []any {
	items := make([]any, len(navItems))
	for i, ni := range navItems {
		items[i] = map[string]any{
			"Path":  ni.Path,
			"Label": ni.Label,
			"Meta":  ni.Meta,
		}
	}
	return items
}

// RouterData returns a map of router-exposed template variables.
// Intended to be merged into component Props automatically.
func RouterData() map[string]any {
	return map[string]any{
		"ActivePath": activePathSig,
		"NavItems":   NavItemsMap(),
	}
}

// ActivePath returns the reactive signal holding the current route path.
func ActivePath() *state.Signal[string] {
	return activePathSig
}
