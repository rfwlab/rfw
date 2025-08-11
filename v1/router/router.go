//go:build js && wasm

package router

import (
	"log"
	"regexp"
	"strings"
	"syscall/js"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
)

type route struct {
	pattern    string
	regex      *regexp.Regexp
	paramNames []string
	component  core.Component
}

var routes []route
var currentComponent core.Component

func RegisterRoute(path string, component core.Component) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
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

	pattern := "^/" + strings.Join(regexParts, "/") + "$"
	r := route{
		pattern:    path,
		regex:      regexp.MustCompile(pattern),
		paramNames: paramNames,
		component:  component,
	}
	routes = append(routes, r)
}

type routeParamReceiver interface {
	SetRouteParams(map[string]string)
}

func Navigate(path string) {
	for _, r := range routes {
		if matches := r.regex.FindStringSubmatch(path); matches != nil {
			params := map[string]string{}
			for i, name := range r.paramNames {
				if i+1 < len(matches) {
					params[name] = matches[i+1]
				}
			}
			if receiver, ok := r.component.(routeParamReceiver); ok {
				receiver.SetRouteParams(params)
			}
			if currentComponent != nil {
				log.Println("Unmounting current component:", currentComponent.GetName())
				currentComponent.Unmount()
			}
			currentComponent = r.component
			dom.UpdateDOM("", r.component.Render())
			js.Global().Get("history").Call("pushState", nil, "", path)
			return
		}
	}
}

func ExposeNavigate() {
	js.Global().Set("goNavigate", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		path := args[0].String()
		Navigate(path)
		return nil
	}))
}

func InitRouter() {
	js.Global().Get("window").Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		path := js.Global().Get("location").Get("pathname").String()
		Navigate(path)
		return nil
	}))

	currentPath := js.Global().Get("location").Get("pathname").String()
	Navigate(currentPath)
}
