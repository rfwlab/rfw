package router

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/rfwlab/rfw/v2/state"
)

var (
	// ErrRouteNotFound reports a path with no registered destination.
	ErrRouteNotFound = errors.New("router: route not found")
	// ErrNavigationBlocked reports a guard that rejected a destination.
	ErrNavigationBlocked = errors.New("router: navigation blocked")
	// ErrRouteComponent reports a matched route without a component.
	ErrRouteComponent = errors.New("router: route has no component")
	// ErrRedirectLoop reports too many consecutive route redirects.
	ErrRedirectLoop = errors.New("router: redirect loop")
)

// LoadContext describes the destination passed to a route loader.
type LoadContext struct {
	Path   string
	Params map[string]string
	Query  url.Values
}

// Loader resolves data before a route component is committed.
type Loader func(context.Context, LoadContext) (any, error)

// RouteDataReceiver accepts data returned by a route Loader.
type RouteDataReceiver interface {
	SetRouteData(any)
}

// NavigationStatus describes the active router transition.
type NavigationStatus string

const (
	NavigationIdle    NavigationStatus = "idle"
	NavigationLoading NavigationStatus = "loading"
	NavigationReady   NavigationStatus = "ready"
	NavigationError   NavigationStatus = "error"
)

var (
	navigationStatus = state.NewSignal(NavigationIdle)
	navigationError  = state.NewSignal[error](nil)
	currentRouteData = state.NewSignal[any](nil)
	currentRouteMeta = state.NewSignal(map[string]any{})

	navigationMu     sync.Mutex
	navigationID     uint64
	navigationCancel context.CancelFunc
)

// Status returns the reactive navigation status.
func Status() *state.Signal[NavigationStatus] {
	return navigationStatus
}

// Error returns the latest reactive loader error.
func Error() *state.Signal[error] {
	return navigationError
}

// Data returns the current route's reactive loader data.
func Data() *state.Signal[any] {
	return currentRouteData
}

// Meta returns the current route's reactive metadata.
func Meta() *state.Signal[map[string]any] {
	return currentRouteMeta
}

func beginNavigation(parent context.Context) (context.Context, uint64) {
	if parent == nil {
		parent = context.Background()
	}
	navigationMu.Lock()
	if navigationCancel != nil {
		navigationCancel()
	}
	ctx, cancel := context.WithCancel(parent)
	navigationCancel = cancel
	navigationID++
	id := navigationID
	navigationMu.Unlock()
	return ctx, id
}

func navigationIsCurrent(id uint64) bool {
	navigationMu.Lock()
	current := navigationID == id
	navigationMu.Unlock()
	return current
}

func resetNavigation() {
	navigationMu.Lock()
	if navigationCancel != nil {
		navigationCancel()
	}
	navigationCancel = nil
	navigationID++
	navigationMu.Unlock()
	state.Batch(func() {
		navigationStatus.Set(NavigationIdle)
		navigationError.Set(nil)
		currentRouteData.Set(nil)
		currentRouteMeta.Set(map[string]any{})
	})
}

func commitRouteState(data any, meta map[string]any) {
	state.Batch(func() {
		currentRouteData.Set(data)
		currentRouteMeta.Set(cloneMeta(meta))
		navigationError.Set(nil)
		navigationStatus.Set(NavigationReady)
	})
}

func failNavigation(err error) {
	state.Batch(func() {
		navigationError.Set(err)
		navigationStatus.Set(NavigationError)
	})
}

func cloneMeta(meta map[string]any) map[string]any {
	if meta == nil {
		return map[string]any{}
	}
	clone := make(map[string]any, len(meta))
	for key, value := range meta {
		clone[key] = value
	}
	return clone
}

func cloneStringMap(values map[string]string) map[string]string {
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

// URL builds a URL for a named route.
func URL(name string, params map[string]string, query url.Values) (string, error) {
	template, ok := namedRoutePath(routes, name)
	if !ok {
		return "", errors.New("router: named route not found")
	}
	segments := strings.Split(template, "/")
	for index, segment := range segments {
		if !strings.HasPrefix(segment, ":") {
			continue
		}
		key := strings.TrimPrefix(segment, ":")
		value, exists := params[key]
		if !exists || value == "" {
			return "", errors.New("router: missing route parameter " + key)
		}
		segments[index] = url.PathEscape(value)
	}
	path := strings.Join(segments, "/")
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return path, nil
}

// MustURL builds a named URL and panics if it is invalid.
func MustURL(name string, params map[string]string, query url.Values) string {
	path, err := URL(name, params, query)
	if err != nil {
		panic(err)
	}
	return path
}

func namedRoutePath(list []route, name string) (string, bool) {
	for index := range list {
		if list[index].name == name {
			return list[index].fullPath, true
		}
		if path, ok := namedRoutePath(list[index].children, name); ok {
			return path, true
		}
	}
	return "", false
}

func routeQuery(raw string) (map[string]string, url.Values) {
	values, _ := url.ParseQuery(raw)
	params := make(map[string]string, len(values))
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if entries := values[key]; len(entries) > 0 {
			params[key] = entries[0]
		}
	}
	return params, values
}

type redirectDepthKey struct{}

func nextRedirectContext(parent context.Context) (context.Context, error) {
	depth, _ := parent.Value(redirectDepthKey{}).(int)
	if depth >= 16 {
		return nil, ErrRedirectLoop
	}
	return context.WithValue(parent, redirectDepthKey{}, depth+1), nil
}

func redirectPath(template string, params map[string]string) (string, error) {
	segments := strings.Split(template, "/")
	for index, segment := range segments {
		if !strings.HasPrefix(segment, ":") {
			continue
		}
		key := strings.TrimPrefix(segment, ":")
		value, ok := params[key]
		if !ok || value == "" {
			return "", errors.New("router: missing redirect parameter " + key)
		}
		segments[index] = url.PathEscape(value)
	}
	return strings.Join(segments, "/"), nil
}
