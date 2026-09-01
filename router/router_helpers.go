package router

import "github.com/rfwlab/rfw/v2/core"

// NavigateTo is an alias for Navigate.
func NavigateTo(fullPath string) { Navigate(fullPath) }

// CurrentComponent returns the current routed component.
func CurrentComponent() core.Component { return currentComponent }
