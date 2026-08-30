// Package router provides client-side application routing.
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
	// ErrNavigationForbidden reports a result guard that refused a destination
	// outright, with no alternative route to send the caller to.
	ErrNavigationForbidden = errors.New("router: navigation forbidden")
	// ErrInvalidGuardResult reports a result guard that returned an unknown
	// action or a redirect without a usable destination.
	ErrInvalidGuardResult = errors.New("router: invalid guard result")
	// ErrHostNavigation reports a host handoff that the current build cannot
	// perform. Browser builds replace the document instead of returning it.
	ErrHostNavigation = errors.New("router: host navigation requested")
)

// HostNavigationError carries the host path a guard handed navigation to.
// Outside browser builds there is no document to replace, so the router
// reports the request rather than pretending the host page loaded. It wraps
// ErrHostNavigation, so errors.Is identifies it without a type assertion.
type HostNavigationError struct {
	// Path is the validated rooted path the host was asked to serve.
	Path string
}

// Error reports the handoff and the path it targets.
func (e *HostNavigationError) Error() string {
	return ErrHostNavigation.Error() + ": " + e.Path
}

// Unwrap returns ErrHostNavigation, the sentinel callers match on.
func (e *HostNavigationError) Unwrap() error { return ErrHostNavigation }

// GuardAction is the outcome a result guard asks the router to apply.
type GuardAction uint8

const (
	// GuardAllow continues with the next guard and, eventually, the route.
	GuardAllow GuardAction = iota
	// GuardForbid stops navigation with ErrNavigationForbidden.
	GuardForbid
	// GuardRedirect navigates to another path and keeps a history entry for
	// the current one, except on an initial load or a popstate, where the
	// denied entry is replaced instead.
	GuardRedirect
	// GuardReplace navigates to another path in place of the current history
	// entry, which is what a login redirect wants.
	GuardReplace
	// GuardHostReplace leaves the application: the browser loads the path from
	// the host, so the router matches no route and commits nothing.
	GuardHostReplace
)

// GuardResult is the decision a ResultGuard returns. Build it with Allow,
// Forbid, RedirectTo, ReplaceWith or HostReplace rather than by hand.
type GuardResult struct {
	Action GuardAction
	Path   string
}

// ResultGuard decides whether a route may be entered and, when it may not,
// where navigation goes instead. A result guard must not call Navigate or
// Replace itself: returning the destination lets the router apply it once, in
// the right order, without reentering an in-flight navigation.
type ResultGuard func(map[string]string) GuardResult

// Allow lets navigation continue.
func Allow() GuardResult { return GuardResult{Action: GuardAllow} }

// Forbid refuses navigation without a destination to fall back to.
func Forbid() GuardResult { return GuardResult{Action: GuardForbid} }

// RedirectTo sends navigation to path and keeps the current history entry,
// unless navigation came from an initial load or a popstate.
func RedirectTo(path string) GuardResult {
	return GuardResult{Action: GuardRedirect, Path: path}
}

// ReplaceWith sends navigation to path in place of the current history entry.
func ReplaceWith(path string) GuardResult {
	return GuardResult{Action: GuardReplace, Path: path}
}

// HostReplace hands the document to a host-owned path on the same origin, the
// server-rendered login or session-expiry page an application does not route
// itself. The browser loads it as a real navigation and replaces the current
// history entry; RedirectTo and ReplaceWith stay inside the SPA and only
// target registered routes. The path must already be rooted and same origin as
// written, surrounding whitespace included: an external or ambiguous target
// fails closed with ErrInvalidGuardResult.
func HostReplace(path string) GuardResult {
	return GuardResult{Action: GuardHostReplace, Path: path}
}

// guardEntry is one guard in evaluation order. Exactly one of its fields is
// set: legacy guards keep their allow-or-block contract, result guards carry
// the typed one.
type guardEntry struct {
	legacy Guard
	result ResultGuard
}

// routeGuardEntries orders the guards of a single route: its legacy guards
// first, then its result guards. Nesting keeps parents before children, so the
// whole chain is deterministic.
func routeGuardEntries(legacy []Guard, results []ResultGuard) []guardEntry {
	if len(legacy) == 0 && len(results) == 0 {
		return nil
	}
	entries := make([]guardEntry, 0, len(legacy)+len(results))
	for _, guard := range legacy {
		entries = append(entries, guardEntry{legacy: guard})
	}
	for _, guard := range results {
		entries = append(entries, guardEntry{result: guard})
	}
	return entries
}

// joinGuards concatenates a parent chain with its child chain without writing
// into the parent's own slice.
func joinGuards(parent, child []guardEntry) []guardEntry {
	if len(parent) == 0 {
		return child
	}
	if len(child) == 0 {
		return parent
	}
	joined := make([]guardEntry, 0, len(parent)+len(child))
	joined = append(joined, parent...)
	return append(joined, child...)
}

// runGuards evaluates the chain and stops at the first guard that does not
// allow navigation. The returned result is meaningful only when err is nil: a
// GuardRedirect, GuardReplace or GuardHostReplace action carries the validated
// destination, and GuardAllow means the route may be loaded.
func runGuards(entries []guardEntry, params map[string]string) (GuardResult, error) {
	for _, entry := range entries {
		if entry.legacy != nil {
			if entry.legacy(params) {
				continue
			}
			return GuardResult{}, ErrNavigationBlocked
		}
		if entry.result == nil {
			continue
		}
		result := entry.result(params)
		switch result.Action {
		case GuardAllow:
			continue
		case GuardForbid:
			return GuardResult{}, ErrNavigationForbidden
		case GuardRedirect, GuardReplace:
			destination, err := guardRedirectTarget(result.Path)
			if err != nil {
				return GuardResult{}, err
			}
			result.Path = destination
			return result, nil
		case GuardHostReplace:
			destination, err := guardHostReplaceTarget(result.Path)
			if err != nil {
				return GuardResult{}, err
			}
			result.Path = destination
			return result, nil
		default:
			return GuardResult{}, ErrInvalidGuardResult
		}
	}
	return GuardResult{}, nil
}

// guardRedirectTarget validates a guard destination. It must be a rooted
// application path: an empty, relative or scheme-relative target would send
// navigation somewhere the router cannot resolve, so it fails closed.
func guardRedirectTarget(path string) (string, error) {
	destination := strings.TrimSpace(path)
	if destination == "" || !strings.HasPrefix(destination, "/") {
		return "", ErrInvalidGuardResult
	}
	// "//host" and "/\host" leave the application: browsers read them as
	// protocol-relative URLs, which is an open redirect, not a route.
	if strings.HasPrefix(destination, "//") || strings.HasPrefix(destination, `/\`) {
		return "", ErrInvalidGuardResult
	}
	return destination, nil
}

// guardHostReplaceTarget validates a host handoff destination. This one is not
// matched against a route: the browser loads it, so what matters is the URL the
// browser ends up with after its own normalization. The string a guard returned
// is therefore checked as it is and never repaired first: trimming would accept
// " /login" by turning it into a different string, and the caller would have
// handed the browser something validation never saw. A target that is not
// already a rooted same-origin path fails closed, and an accepted one is
// returned byte for byte.
func guardHostReplaceTarget(path string) (string, error) {
	if path == "" || !strings.HasPrefix(path, "/") {
		return "", ErrInvalidGuardResult
	}
	// Browsers remove tabs and newlines from anywhere in a URL, so "/<tab>/host"
	// reaches the network as "//host", and they strip C0 characters and spaces
	// from both ends before parsing. Refuse them rather than validate a string
	// the browser will rewrite: a leading one already fails the rooted check
	// above, and a trailing space is the only stripped byte left to catch.
	if strings.ContainsFunc(path, isURLControl) {
		return "", ErrInvalidGuardResult
	}
	if strings.HasSuffix(path, " ") {
		return "", ErrInvalidGuardResult
	}
	// A backslash is read as a slash, which turns "/\host" into a
	// protocol-relative URL pointing somewhere else entirely.
	if strings.ContainsRune(path, '\\') {
		return "", ErrInvalidGuardResult
	}
	if strings.HasPrefix(path, "//") {
		return "", ErrInvalidGuardResult
	}
	parsed, err := url.Parse(path)
	if err != nil {
		return "", ErrInvalidGuardResult
	}
	if parsed.Scheme != "" || parsed.Opaque != "" || parsed.Host != "" || parsed.User != nil {
		return "", ErrInvalidGuardResult
	}
	if !strings.HasPrefix(parsed.Path, "/") {
		return "", ErrInvalidGuardResult
	}
	// Parse keeps the query raw, so a malformed escape there survives it.
	if _, err := url.ParseQuery(parsed.RawQuery); err != nil {
		return "", ErrInvalidGuardResult
	}
	return path, nil
}

func isURLControl(r rune) bool { return r < 0x20 || r == 0x7f }

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
	// NavigationIdle indicates that no navigation is active.
	NavigationIdle NavigationStatus = "idle"
	// NavigationLoading indicates that a route loader is running.
	NavigationLoading NavigationStatus = "loading"
	// NavigationReady indicates that navigation completed successfully.
	NavigationReady NavigationStatus = "ready"
	// NavigationError indicates that navigation failed.
	NavigationError NavigationStatus = "error"
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
