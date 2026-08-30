package router

import (
	"context"
	"errors"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
)

// hostReplaceCase names a host handoff target under test.
type hostReplaceCase struct {
	name   string
	target string
}

// hostReplaceRejects lists targets a guard must never be able to hand the
// browser. Each one either leaves the origin outright or is rewritten into
// something that does by the normalization a browser applies before loading a
// URL, so validation has to refuse the string it was given rather than repair
// it into one it would accept.
var hostReplaceRejects = []hostReplaceCase{
	{"empty", ""},
	{"blank", "   "},
	{"leading space", " /login"},
	{"trailing space", "/login "},
	{"surrounding spaces", "  /login  "},
	{"leading tab", "\t/login"},
	{"trailing tab", "/login\t"},
	{"leading carriage return", "\r/login"},
	{"trailing crlf", "/login\r\n"},
	{"surrounding whitespace", "  /login\r\n"},
	{"relative", "login"},
	{"relative dot", "./login"},
	{"absolute scheme", "https://evil.example/login"},
	{"scheme relative", "//evil.example/login"},
	{"newline scheme relative", "\n//evil.example/login"},
	{"backslash authority", `/\evil.example/login`},
	{"double backslash", `/\\evil.example`},
	{"backslash in path", `/login\..\evil`},
	{"backslash in query", `/login?next=/\evil.example`},
	{"tab authority", "/\t/evil.example"},
	{"embedded newline", "/lo\ngin"},
	{"embedded carriage return", "/lo\r\ngin"},
	{"embedded tab", "/log\tin"},
	{"vertical tab", "/log\vin"},
	{"null byte", "/login\x00"},
	{"delete", "/log\x7fin"},
	{"opaque scheme", "mailto:ops@evil.example"},
	{"javascript scheme", "javascript:alert(1)"},
	{"scheme backslash", `https:/\evil.example`},
	{"truncated percent", "/login%"},
	{"malformed percent path", "/log%2in"},
	{"malformed percent query", "/login?next=%zz"},
	{"malformed percent fragment", "/login#%zz"},
}

// hostReplaceAccepts lists the ordinary host-owned paths an application hands
// off to: a rooted path, with or without a query and a fragment.
var hostReplaceAccepts = []string{
	"/",
	"/login",
	"/login?next=%2Fadmin",
	"/settings#billing",
	"/login?next=%2Fadmin#top",
	"/account/session-expired",
}

// sessionGuardedRoute registers /composed behind the guard an authenticated
// application writes: no session hands the browser to the host login page, a
// session without the permission is refused inside the SPA, and anything else
// enters normally.
func sessionGuardedRoute(t *testing.T, authenticated, permitted *bool) *int {
	t.Helper()
	created := 0
	RegisterRoute(Route{
		Path: "/composed",
		ResultGuards: []ResultGuard{func(map[string]string) GuardResult {
			switch {
			case !*authenticated:
				return HostReplace("/login")
			case !*permitted:
				return Forbid()
			default:
				return Allow()
			}
		}},
		Component: func() core.Component {
			created++
			return &recordComponent{name: "composed"}
		},
	})
	return &created
}

func TestHostReplaceTargetRejectsOffOriginTargets(t *testing.T) {
	for _, tc := range hostReplaceRejects {
		t.Run(tc.name, func(t *testing.T) {
			destination, err := guardHostReplaceTarget(tc.target)
			if !errors.Is(err, ErrInvalidGuardResult) {
				t.Fatalf("guardHostReplaceTarget(%q) = %q, %v", tc.target, destination, err)
			}
			if destination != "" {
				t.Fatalf("rejected target returned a destination: %q", destination)
			}
		})
	}
}

func TestHostReplaceTargetAcceptsRootedPaths(t *testing.T) {
	for _, target := range hostReplaceAccepts {
		t.Run(target, func(t *testing.T) {
			destination, err := guardHostReplaceTarget(target)
			if err != nil {
				t.Fatalf("guardHostReplaceTarget(%q): %v", target, err)
			}
			if destination != target {
				t.Fatalf("destination = %q, want %q", destination, target)
			}
		})
	}
}

// A host handoff is its own outcome: it is neither a refusal nor a route the
// router could resolve, and it carries the target it was asked for.
func TestHostNavigationErrorIsDistinct(t *testing.T) {
	var err error = &HostNavigationError{Path: "/login"}
	if !errors.Is(err, ErrHostNavigation) {
		t.Fatalf("host navigation error does not wrap ErrHostNavigation: %v", err)
	}
	if errors.Is(err, ErrNavigationForbidden) || errors.Is(err, ErrRouteNotFound) || errors.Is(err, ErrInvalidGuardResult) {
		t.Fatalf("host navigation error matches another outcome: %v", err)
	}
	var host *HostNavigationError
	if !errors.As(err, &host) || host.Path != "/login" {
		t.Fatalf("host navigation error target = %#v", host)
	}
	if err.Error() != "router: host navigation requested: /login" {
		t.Fatalf("error message = %q", err.Error())
	}
}

// The SPA redirects keep their own validation: a host handoff loosens nothing
// for RedirectTo and ReplaceWith.
func TestGuardRedirectTargetStillRejectsOffOriginTargets(t *testing.T) {
	for _, target := range []string{"", "  ", "login", "//evil.example", `/\evil.example`} {
		if _, err := guardRedirectTarget(target); !errors.Is(err, ErrInvalidGuardResult) {
			t.Fatalf("guardRedirectTarget(%q) = %v", target, err)
		}
	}
}

// Allow and Forbid are unchanged by the new action: an allowed navigation
// enters the route, and a refused one stays inside the SPA with
// ErrNavigationForbidden instead of leaving for the host.
func TestSessionGuardAllowsAndForbidsInsideSPA(t *testing.T) {
	resetRouter(t)
	authenticated, permitted := true, true
	created := sessionGuardedRoute(t, &authenticated, &permitted)

	if err := NavigateContext(context.Background(), "/composed"); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	if got := mustRecord(t, CurrentComponent()).name; got != "composed" {
		t.Fatalf("current component = %q, want composed", got)
	}
	if *created != 1 {
		t.Fatalf("component created %d times, want 1", *created)
	}

	permitted = false
	err := NavigateContext(context.Background(), "/composed")

	if !errors.Is(err, ErrNavigationForbidden) {
		t.Fatalf("expected a forbidden navigation, got %v", err)
	}
	if errors.Is(err, ErrHostNavigation) {
		t.Fatal("an authenticated caller without the permission was sent to the host")
	}
	if *created != 1 {
		t.Fatalf("forbidden navigation created the component again: %d", *created)
	}
}
