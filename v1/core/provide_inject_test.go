//go:build js && wasm

package core

import (
	"testing"

	"github.com/rfwlab/rfw/v1/state"
)

func TestProvideInject(t *testing.T) {
	state.NewStore("default", state.WithModule("app"))

	parentTpl := []byte("<root></root>")
	childTpl := []byte("<root></root>")

	parent := NewComponent("Parent", parentTpl, nil)
	child := NewComponent("Child", childTpl, nil)

	parent.Provide("answer", 42)
	parent.AddDependency("child", child)

	v, ok := Inject[int](child, "answer")
	if !ok || v != 42 {
		t.Fatalf("expected injected 42, got %v", v)
	}
}
