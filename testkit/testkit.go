// Package testkit provides small helpers for component and asynchronous tests.
package testkit

import (
	"strings"
	"time"

	"github.com/rfwlab/rfw/v2/core"
)

// TestingT is the subset of testing.T used by this package.
type TestingT interface {
	Helper()
	Fatalf(string, ...any)
	Cleanup(func())
}

// Render returns a component's rendered HTML without mounting it.
func Render(component core.Component) string {
	if component == nil {
		return ""
	}
	return component.Render()
}

// AssertContains fails when text does not contain expected.
func AssertContains(t TestingT, text, expected string) {
	t.Helper()
	if !strings.Contains(text, expected) {
		t.Fatalf("expected %q to contain %q", text, expected)
	}
}

// Eventually retries condition until it succeeds or timeout expires.
func Eventually(t TestingT, timeout time.Duration, condition func() bool) {
	t.Helper()
	if timeout <= 0 {
		timeout = time.Second
	}
	deadline := time.Now().Add(timeout)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("condition was not met within %s", timeout)
		}
		time.Sleep(time.Millisecond)
	}
}
