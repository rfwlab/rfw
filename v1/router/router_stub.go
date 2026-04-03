//go:build !js || !wasm

// Package router provides a headless router implementation for non-JS/WASM
// builds. It mirrors the route matching and guard behavior of the JS/WASM
// router but skips DOM/History integration.
package router

import (
"net/url"
"regexp"
"strings"

"github.com/rfwlab/rfw/v1/core"
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

// RegisteredRoute describes a registered route in a navigable tree form.
type RegisteredRoute struct {
Template string            `json:"template"`
Path     string            `json:"path"`
Params   []string          `json:"params"`
Children []RegisteredRoute `json:"children"`
}

var (
routes           []route
currentComponent core.Component
)

// NotFoundComponent, if set, is rendered when no route matches the path.
var NotFoundComponent func() core.Component

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
}

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

pathRegex := strings.Join(regexParts, "/")
suffix := "/?$"
if len(r.Children) > 0 {
suffix = "(?:/|$)"
}
if pathRegex == "" {
suffix = "$"
}
pattern := "^/" + pathRegex + suffix

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

// RegisteredRoutes returns the registered routes including nested children and
// resolved full paths.
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

// Navigate routes to fullPath if it matches a registered route and all guards
// allow it.
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
currentComponent = NotFoundComponent()
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

currentComponent = r.component
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

// ExposeNavigate is a no-op in non-WASM builds.
func ExposeNavigate() {}

// InitRouter is a no-op in non-WASM builds.
func InitRouter() {}
