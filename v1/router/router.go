//go:build js && wasm

// Package router implements a simple client-side router for WebAssembly
// applications built with RFW.
package router

import (
	"net/url"
	"regexp"
	"strings"
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
	events "github.com/rfwlab/rfw/v1/events"
	"github.com/rfwlab/rfw/v1/js"
)

// Guard is a function that determines whether navigation to a route is
// permitted based on the provided parameters.
type Guard func(map[string]string) bool

// Route describes a routing rule that maps a path to a component and optional
// guards or child routes.
type Route struct {
	Path      string
	Component func() core.Component
	Guards    []Guard
	Children  []Route
}

type route struct {
	pattern    string
	regex      *regexp.Regexp
	paramNames []string
	component  core.Component
	loader     func() core.Component
	children   []route
	guards     []Guard
}

var routes []route
var currentComponent core.Component

// NotFoundComponent, if set, is rendered when no route matches the path.
var NotFoundComponent func() core.Component

// NotFoundCallback, if set, is invoked when navigation targets an
// unregistered route. It receives the requested path.
var NotFoundCallback func(string)

// RegisterRoute adds a new Route to the router's configuration.
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

	suffix := "$"
	if len(r.Children) > 0 {
		suffix = "(?:/|$)"
	}
	pattern := "^/" + strings.Join(regexParts, "/") + suffix
	rt := route{
		pattern:    r.Path,
		regex:      regexp.MustCompile(pattern),
		paramNames: paramNames,
		loader:     r.Component,
		guards:     r.Guards,
	}

	for _, child := range r.Children {
		rt.children = append(rt.children, buildRoute(child))
	}

	return rt
}

type routeParamReceiver interface {
	SetRouteParams(map[string]string)
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

// Navigate renders the component associated with the specified path if all
// route guards allow it. The provided path may include a query string which
// will be parsed and passed to the component via SetRouteParams.
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
			if currentComponent != nil {
				core.Log().Debug("Unmounting current component: %s", currentComponent.GetName())
				core.TriggerUnmount(currentComponent)
				currentComponent.Unmount()
			}
			c := NotFoundComponent()
			currentComponent = c
			dom.UpdateDOM("", c.Render())
			c.Mount()
			core.TriggerMount(c)
			core.TriggerRouter(fullPath)
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

	if r.component == nil && r.loader != nil {
		r.component = r.loader()
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
	if currentComponent != nil {
		core.Log().Debug("Unmounting current component: %s", currentComponent.GetName())
		core.TriggerUnmount(currentComponent)
		currentComponent.Unmount()
	}
	currentComponent = r.component
	dom.UpdateDOM("", r.component.Render())
	r.component.Mount()
	core.TriggerMount(r.component)
	core.TriggerRouter(fullPath)
	js.History().Call("pushState", nil, "", fullPath)
}

// ExposeNavigate makes the Navigate function accessible from JavaScript.
func ExposeNavigate() {
	js.ExposeFunc("goNavigate", func(this jst.Value, args []jst.Value) any {
		path := args[0].String()
		Navigate(path)
		return nil
	})
}

// InitRouter initializes the router and begins listening for navigation
// events.
func InitRouter() {
	ch := events.Listen("popstate", js.Window())
	go func() {
		for range ch {
			path := js.Location().Get("pathname").String() + js.Location().Get("search").String()
			Navigate(path)
		}
	}()

	currentPath := js.Location().Get("pathname").String() + js.Location().Get("search").String()
	Navigate(currentPath)
}
