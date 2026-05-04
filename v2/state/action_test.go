package state

import (
	"context"
	"testing"
)

func TestDispatch(t *testing.T) {
	called := false
	a := Action(func(ctx Context) error {
		called = true
		return nil
	})
	if err := Dispatch(context.Background(), a); err != nil {
		t.Fatalf("dispatch returned error: %v", err)
	}
	if !called {
		t.Fatalf("action was not executed")
	}
}

func TestUseAction(t *testing.T) {
	called := false
	a := Action(func(ctx Context) error {
		called = true
		return nil
	})
	fn := UseAction(context.Background(), a)
	if err := fn(); err != nil {
		t.Fatalf("use action returned error: %v", err)
	}
	if !called {
		t.Fatalf("action not executed via UseAction")
	}
}
