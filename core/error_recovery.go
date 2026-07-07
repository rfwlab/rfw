//go:build js && wasm

package core

import (
	"fmt"
	"runtime/debug"

	"github.com/rfwlab/rfw/v2/state"
)

// TryRender wraps a component's Render() with panic recovery.
// If a panic occurs, it shows the error overlay and returns empty string
// so the app stays alive rather than dying to a white screen.
func TryRender(c Component) string {
	defer func() {
		if r := recover(); r != nil {
			ShowErrorOverlay(r, fmt.Sprintf("Render: %s (ID: %s)", c.GetName(), c.GetID()))
		}
	}()
	return c.Render()
}

// TryMount wraps a component's Mount() with panic recovery.
func TryMount(c Component) {
	defer func() {
		if r := recover(); r != nil {
			ShowErrorOverlay(r, fmt.Sprintf("Mount: %s (ID: %s)", c.GetName(), c.GetID()))
		}
	}()
	c.Mount()
}

// TryUnmount wraps a component's Unmount() with panic recovery.
func TryUnmount(c Component) {
	defer func() {
		if r := recover(); r != nil {
			ShowErrorOverlay(r, fmt.Sprintf("Unmount: %s (ID: %s)", c.GetName(), c.GetID()))
		}
	}()
	c.Unmount()
}

// TryNavigate wraps router navigation with panic recovery.
func TryNavigate(path string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			ShowErrorOverlay(r, fmt.Sprintf("Navigate: %s", path))
		}
	}()
	fn()
}

// TryEffect wraps an effect function with panic recovery.
func TryEffect(fn func() func()) func() {
	return state.Effect(func() func() {
		defer func() {
			if r := recover(); r != nil {
				ShowErrorOverlay(r, "Effect")
				debug.PrintStack()
			}
		}()
		return fn()
	})
}

// TryTemplateLoad wraps template loading with recovery.
func TryTemplateLoad(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			ShowErrorOverlay(r, "Template / Composition")
		}
	}()
	fn()
}
