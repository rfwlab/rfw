//go:build !js || !wasm

package router

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/state"
	"github.com/rfwlab/rfw/v2/types"
)

type Guard func(map[string]string) bool

type Route struct {
	Path      string
	Component any
	Guards    []Guard
	Children  []Route
}

func Singleton(v *types.View) any {
	return v
}

type route struct {
	pattern    string
	regex      *regexp.Regexp
	paramNames []string
	component  core.Component
	loader     func() core.Component
	singleton  bool
	children   []route
	guards     []Guard
}

type RegisteredRoute struct {
	Template string            `json:"template"`
	Path     string            `json:"path"`
	Params   []string          `json:"params"`
	Children []RegisteredRoute `json:"children"`
}

var (
	routes           []route
	currentComponent core.Component
	activePathSig    = state.NewSignal("/")
	navItems         []NavItem
)

var NotFoundComponent any
var NotFoundCallback func(string)

func Reset() {
	routes = nil
	currentComponent = nil
	NotFoundComponent = nil
	NotFoundCallback = nil
}

func RegisterRoute(r Route) {
	routes = append(routes, buildRoute(r))
}

func buildRoute(r Route) route {
	segments := strings.Split(strings.Trim(r.Path, "/"), "/")
	regexParts := make([]string, len(segments))
	paramNames := []string{}

	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			name := strings.TrimPrefix(segment, ":")
			paramNames = append(paramNames, name)
			regexParts[i] = "([^/]+)"
		} else {
			regexParts[i] = regexp.QuoteMeta(segment)
		}
	}

	pathRegex := strings.Join(regexParts, "/")
	suffix := "/?$"
	if len(r.Children) > 0 {
		suffix = "(?:/|$)"
	}
	if pathRegex == "" {
		suffix = "$"
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
	}
	rt := route{
		pattern:    r.Path,
		regex:      regexp.MustCompile(pattern),
		paramNames: paramNames,
		loader:     loader,
		singleton:  singleton,
		guards:     r.Guards,
	}

	for _, child := range r.Children {
		rt.children = append(rt.children, buildRoute(child))
	}

	return rt
}

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
		Params:   params,
		Children: children,
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

func matchRoute(routes []route, path string) (*route, []Guard, map[string]string) {
	for i := range routes {
		r := &routes[i]
		if matches := r.regex.FindStringSubmatch(path); matches != nil {
			params := map[string]string{}
			for i, name := range r.paramNames {
				if i+1 < len(matches) {
					params[name] = matches[i+1]
				}
			}
			if child, guards, childParams := matchRoute(r.children, path); child != nil {
				for k, v := range params {
					childParams[k] = v
				}
				return child, append(r.guards, guards...), childParams
			}
			return r, r.guards, params
		}
	}
	return nil, nil, nil
}

func Navigate(fullPath string) {
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
		}
		return
	}

	for _, g := range guards {
		if !g(params) {
			if currentComponent == nil && path != "/" {
				Navigate("/")
			}
			return
		}
	}

	if r.loader != nil {
		if r.singleton && r.component != nil {
			// Re-use existing instance for singleton routes.
		} else {
			r.component = r.loader()
		}
	}

	if params == nil {
		params = map[string]string{}
	}
	if query != "" {
		if values, err := url.ParseQuery(query); err == nil {
			for k, v := range values {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}
		}
	}
	if receiver, ok := r.component.(routeParamReceiver); ok {
		receiver.SetRouteParams(params)
	}

	currentComponent = r.component
	if handler, ok := r.component.(routeParamHandler); ok {
		handler.OnParams(params)
	}
}

func CanNavigate(fullPath string) bool {
	path := fullPath
	if idx := strings.Index(fullPath, "?"); idx != -1 {
		path = fullPath[:idx]
	}
	r, _, _ := matchRoute(routes, path)
	return r != nil
}

func Page(path string, component any, guards ...Guard) {
	RegisterRoute(Route{
		Path:      path,
		Component: component,
		Guards:    guards,
	})
}

func Group(prefix string, fn func(*GroupBuilder)) {
	b := &GroupBuilder{prefix: prefix}
	fn(b)
	RegisterRoute(Route{
		Path:     prefix,
		Children: b.children,
	})
}

type GroupBuilder struct {
	prefix   string
	children []Route
}

func (g *GroupBuilder) Page(path string, component any, guards ...Guard) {
	g.children = append(g.children, Route{
		Path:      path,
		Component: component,
		Guards:    guards,
	})
}

func ExposeNavigate() {}
func InitRouter()     {}

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

// RouterData returns a map of router-exposed template variables.
func RouterData() map[string]any {
	return map[string]any{
		"ActivePath": activePathSig,
		"NavItems":   NavItemsMap(),
	}
}
