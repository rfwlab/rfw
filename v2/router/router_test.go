package router

import (
"reflect"
"testing"

"github.com/rfwlab/rfw/v2/core"
)

type recordComponent struct {
name   string
params map[string]string
}

func (c *recordComponent) Render() string  { return "" }
func (c *recordComponent) GetName() string { return c.name }
func (c *recordComponent) GetID() string   { return c.name }
func (c *recordComponent) SetRouteParams(p map[string]string) {
c.params = map[string]string{}
for k, v := range p {
c.params[k] = v
}
}

func resetRouter(t *testing.T) {
t.Helper()
Reset()
t.Cleanup(func() { Reset() })
}

func mustRecord(t *testing.T, c core.Component) *recordComponent {
t.Helper()
rc, ok := c.(*recordComponent)
if !ok {
t.Fatalf("expected *recordComponent, got %T", c)
}
return rc
}

func TestRegisterRoute_BasicRouting(t *testing.T) {
resetRouter(t)

RegisterRoute(Route{Path: "/a", Component: func() core.Component { return &recordComponent{name: "a"} }})
RegisterRoute(Route{Path: "/users/:id", Component: func() core.Component { return &recordComponent{name: "user"} }})

NavigateTo("/a")
if got := mustRecord(t, CurrentComponent()).name; got != "a" {
t.Fatalf("expected current component 'a', got %q", got)
}

NavigateTo("/users/123")
rc := mustRecord(t, CurrentComponent())
if rc.params["id"] != "123" {
t.Fatalf("expected id=123, got %v", rc.params)
}

var gotPath string
NotFoundCallback = func(p string) { gotPath = p }
NavigateTo("/missing")
if gotPath != "/missing" {
t.Fatalf("expected NotFoundCallback '/missing', got %q", gotPath)
}
}

func TestNavigateTo_CurrentComponent(t *testing.T) {
resetRouter(t)

NavigateTo("/nothing")
if CurrentComponent() != nil {
t.Fatalf("expected nil current component before routes")
}

RegisterRoute(Route{Path: "/home", Component: func() core.Component { return &recordComponent{name: "home"} }})
NavigateTo("/home")
if got := mustRecord(t, CurrentComponent()).name; got != "home" {
t.Fatalf("expected 'home', got %q", got)
}
}

func TestRouteGuards_BlockNavigation(t *testing.T) {
resetRouter(t)

RegisterRoute(Route{Path: "/", Component: func() core.Component { return &recordComponent{name: "root"} }})

var guardParams map[string]string
RegisterRoute(Route{
Path:      "/admin/:id",
Component: func() core.Component { return &recordComponent{name: "admin"} },
Guards: []Guard{func(p map[string]string) bool {
guardParams = map[string]string{}
for k, v := range p {
guardParams[k] = v
}
return false
}},
})

NavigateTo("/admin/42")
if guardParams["id"] != "42" {
t.Fatalf("expected guard to receive id=42, got %v", guardParams)
}
if got := mustRecord(t, CurrentComponent()).name; got != "root" {
t.Fatalf("expected navigation to '/', got %q", got)
}
}

func TestQueryParams(t *testing.T) {
resetRouter(t)

RegisterRoute(Route{Path: "/search/:kind", Component: func() core.Component { return &recordComponent{name: "search"} }})
NavigateTo("/search/books?q=go&lang=en")

rc := mustRecord(t, CurrentComponent())
want := map[string]string{"kind": "books", "q": "go", "lang": "en"}
if !reflect.DeepEqual(rc.params, want) {
t.Fatalf("expected params %v, got %v", want, rc.params)
}
}

func TestNotFoundComponent(t *testing.T) {
resetRouter(t)

RegisterRoute(Route{Path: "/home", Component: func() core.Component { return &recordComponent{name: "home"} }})
NavigateTo("/home")
_ = mustRecord(t, CurrentComponent())

NotFoundComponent = func() core.Component { return &recordComponent{name: "404"} }
NavigateTo("/missing")

nf := mustRecord(t, CurrentComponent())
if nf.name != "404" {
t.Fatalf("expected 404 current component, got %q", nf.name)
}
}

func TestTrailingSlashNormalization(t *testing.T) {
resetRouter(t)

RegisterRoute(Route{Path: "/trail", Component: func() core.Component { return &recordComponent{name: "trail"} }})
NavigateTo("/trail/")
if got := mustRecord(t, CurrentComponent()).name; got != "trail" {
t.Fatalf("expected trailing slash to match, got %q", got)
}

var called bool
NotFoundCallback = func(string) { called = true }
NavigateTo("/trail/extra")
if !called {
t.Fatalf("expected not found for extra segments")
}
}
