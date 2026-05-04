package state

import "context"

// Context is an alias of context.Context used by Actions.
// This allows the API to remain stable if a custom context is needed later.
type Context = context.Context

// Action represents a unit of work executed with a Context.
// It returns an error if the action fails.
type Action func(ctx Context) error

// Dispatch executes the given Action with the provided context.
// If the action is nil it is a no-op and nil is returned.
func Dispatch(ctx Context, a Action) error {
	if a == nil {
		return nil
	}
	return a(ctx)
}

// UseAction binds an Action to a Context and returns a function
// that executes the action when invoked. It can be used in places
// that expect a simple callback.
func UseAction(ctx Context, a Action) func() error {
	return func() error {
		return Dispatch(ctx, a)
	}
}
