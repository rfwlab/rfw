//go:build !js || !wasm

package router

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/state"
	"github.com/rfwlab/rfw/v2/types"
)

// Guard decides whether a route can be entered.
type Guard func(map[string]string) bool

// Route describes a route and its optional children, loader, and metadata.
type Route struct {
	Path      string
	Name      string
	Component any
	Guards    []Guard
	// ResultGuards run after Guards for this route and can redirect or forbid
	// instead of only blocking.
	ResultGuards []ResultGuard
	Children     []Route
	Loader       Loader
	Redirect     string
	Meta         map[string]any
}

// Singleton marks a view instance for reuse between navigations.
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
	guards     []guardEntry
	dataLoader Loader
	redirect   string
	meta       map[string]any
}

// RegisteredRoute is a serializable snapshot of a route.
type RegisteredRoute struct {
	Template string            `json:"template"`
	Path     string            `json:"path"`
	Name     string            `json:"name,omitempty"`
	Params   []string          `json:"params"`
	Children []RegisteredRoute `json:"children"`
	Meta     map[string]any    `json:"meta,omitempty"`
}

var (
	routes           []route
	currentComponent core.Component
	activePathSig    = state.NewSignal("/")
	navItems         []NavItem
)

// NotFoundComponent is rendered when no route matches.
var NotFoundComponent any

// NotFoundCallback handles an unmatched path.
var NotFoundCallback func(string)

// Reset clears registered routes and navigation state.
func Reset() {
	routes = nil
	currentComponent = nil
	NotFoundComponent = nil
	NotFoundCallback = nil
	activePathSig.Set("/")
	resetNavigation()
}

// RegisterRoute adds a route to the router.
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
		guards:     routeGuardEntries(r.Guards, r.ResultGuards),
		dataLoader: r.Loader,
		redirect:   r.Redirect,
		meta:       cloneMeta(r.Meta),
	}

	for _, child := range r.Children {
		rt.children = append(rt.children, buildRouteAt(child, fullPath))
	}

	return rt
}

// RegisteredRoutes returns snapshots of all registered routes.
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

type routeParamHandler interface {
	OnParams(map[string]string)
}

func matchRoute(routes []route, path string) (*route, []guardEntry, map[string]string) {
	for i := range routes {
		r := &routes[i]
		matches := r.regex.FindStringSubmatch(path)
		if matches == nil {
			if child, guards, params := matchRoute(r.children, path); child != nil {
				return child, joinGuards(r.guards, guards), params
			}
			continue
		}
		params := map[string]string{}
		for i, name := range r.matchNames {
			if i+1 < len(matches) {
				params[name] = decodeRouteParam(matches[i+1])
			}
		}
		if child, guards, childParams := matchRoute(r.children, path); child != nil {
			return child, joinGuards(r.guards, guards), childParams
		}
		matchedPath := strings.TrimSuffix(matches[0], "/")
		if (r.loader != nil || r.redirect != "") && matchedPath == strings.TrimSuffix(path, "/") {
			return r, r.guards, params
		}
	}
	return nil, nil, nil
}

func decodeRouteParam(value string) string {
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value
	}
	return decoded
}

// Navigate loads and commits a route.
func Navigate(fullPath string) {
	_ = NavigateContext(context.Background(), fullPath)
}

// NavigateContext loads and commits a route or returns a navigation error.
func NavigateContext(parent context.Context, fullPath string) error {
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
			switch nc := NotFoundComponent.(type) {
			case *types.View:
				currentComponent = nc
			case func() *types.View:
				currentComponent = nc()
			case func() core.Component:
				currentComponent = nc()
			}
			activePathSig.Set(fullPath)
			commitRouteState(nil, nil)
			return nil
		}
		err := ErrRouteNotFound
		failNavigation(err)
		return err
	}

	if params == nil {
		params = map[string]string{}
	}
	queryParams, queryValues := routeQuery(query)
	for key, value := range queryParams {
		params[key] = value
	}
	// Guards decide before anything commits: no loader, no component and no
	// route state change runs for a destination they refuse. The stub keeps no
	// browser history, so a push and a replace redirect both navigate.
	guardResult, guardErr := runGuards(guards, params)
	if guardErr != nil {
		if guardErr == ErrNavigationBlocked {
			if currentComponent == nil && path != "/" {
				Navigate("/")
			}
			return guardErr
		}
		failNavigation(guardErr)
		return guardErr
	}
	if guardResult.Action == GuardHostReplace {
		// There is no document to replace here, and the host page is not a
		// route this router could load in its place. Report the handoff with
		// its target and leave the current component and route state alone.
		err := &HostNavigationError{Path: guardResult.Path}
		failNavigation(err)
		return err
	}
	if guardResult.Action != GuardAllow {
		redirectCtx, err := nextRedirectContext(parent)
		if err != nil {
			failNavigation(err)
			return err
		}
		return NavigateContext(redirectCtx, guardResult.Path)
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
		return NavigateContext(redirectCtx, destination)
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

	if r.loader != nil && (!r.singleton || r.component == nil) {
		r.component = r.loader()
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

	currentComponent = r.component
	if handler, ok := r.component.(routeParamHandler); ok {
		handler.OnParams(params)
	}
	activePathSig.Set(fullPath)
	commitRouteState(data, r.meta)
	return nil
}

// Replace behaves like Navigate outside browser builds.
func Replace(fullPath string) {
	Navigate(fullPath)
}

// SetScrollRestoration is a no-op outside browser builds.
func SetScrollRestoration(bool) {}

// CanNavigate reports whether a path matches a registered route.
func CanNavigate(fullPath string) bool {
	path := fullPath
	if idx := strings.Index(fullPath, "?"); idx != -1 {
		path = fullPath[:idx]
	}
	r, _, _ := matchRoute(routes, path)
	return r != nil
}

// Page registers a route with an optional set of guards.
func Page(path string, component any, guards ...Guard) {
	RegisterRoute(Route{
		Path:      path,
		Component: component,
		Guards:    guards,
	})
}

// Group registers a set of child routes below a path prefix.
func Group(prefix string, fn func(*GroupBuilder)) {
	b := &GroupBuilder{prefix: prefix}
	fn(b)
	RegisterRoute(Route{
		Path:     prefix,
		Children: b.children,
	})
}

// GroupBuilder collects routes for Group.
type GroupBuilder struct {
	prefix   string
	children []Route
}

// Page adds a route to the group.
func (g *GroupBuilder) Page(path string, component any, guards ...Guard) {
	g.children = append(g.children, Route{
		Path:      path,
		Component: component,
		Guards:    guards,
	})
}

// ExposeNavigate is a no-op outside browser builds.
func ExposeNavigate() {}

// InitRouter is a no-op outside browser builds.
func InitRouter() {}

// NavItem describes a navigation entry with arbitrary metadata.
type NavItem struct {
	Path  string         `json:"path"`
	Label string         `json:"label"`
	Meta  map[string]any `json:"meta"`
}

// SetNavItems registers the navigation items.
func SetNavItems(items []NavItem) {
	navItems = items
}

// NavItems returns the previously registered navigation items.
func NavItems() []NavItem {
	return navItems
}

// NavItemsMap returns navigation items as []any of map[string]any.
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

// RouterData returns the values exposed to component templates.
func RouterData() map[string]any {
	return map[string]any{
		"ActivePath": activePathSig,
		"NavItems":   NavItemsMap(),
	}
}

// TemplateData returns the values exposed to component templates.
func TemplateData() map[string]any { return RouterData() }

// ActivePath returns the reactive signal holding the current route path.
func ActivePath() *state.Signal[string] {
	return activePathSig
}
