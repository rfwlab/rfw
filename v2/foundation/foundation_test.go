package foundation

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestContainerScope(t *testing.T) {
	root := NewContainer()
	root.Provide("name", "root")
	child := root.Scope()
	child.Provide("local", 7)

	if got, ok := child.Get("name"); !ok || got != "root" {
		t.Fatalf("expected scoped container to resolve parent value, got %v %v", got, ok)
	}
	if got, ok := child.Get("local"); !ok || got != 7 {
		t.Fatalf("expected scoped local value, got %v %v", got, ok)
	}
}

func TestEventsAndLifecycle(t *testing.T) {
	bus := &EventBus{bus: DefaultBus.bus}
	type event struct{ Value int }
	var seen int
	Subscribe[event](bus, func(ctx context.Context, e event) error {
		seen = e.Value
		return nil
	})
	if err := Emit[event](bus, event{Value: 3}); err != nil {
		t.Fatalf("emit failed: %v", err)
	}
	if seen != 3 {
		t.Fatalf("expected event value 3, got %d", seen)
	}

	lc := NewLifecycle()
	order := []string{}
	lc.Before("mount", func(ctx context.Context, key string, args []any) error {
		order = append(order, "before:"+key)
		return nil
	})
	lc.After("mount", func(ctx context.Context, key string, args []any) error {
		order = append(order, "after:"+key)
		return nil
	})
	if err := lc.Run(context.Background(), "mount", func() error {
		order = append(order, "action")
		return nil
	}); err != nil {
		t.Fatalf("lifecycle run failed: %v", err)
	}
	want := []string{"before:mount", "action", "after:mount"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("expected lifecycle order %v, got %v", want, order)
	}
}

func TestEffectPipelineAndResult(t *testing.T) {
	ep := NewEffectPipeline()
	called := false
	ep.Use(func(ctx context.Context, input EffectInput, next func(context.Context, EffectInput) (struct{}, error)) (struct{}, error) {
		called = input.ComponentID == "cmp" && input.SignalName == "count" && input.Value == 1
		return next(ctx, input)
	})
	ep.Process(context.Background(), EffectInput{ComponentID: "cmp", SignalName: "count", Value: 1})
	if !called {
		t.Fatalf("expected middleware to observe effect input")
	}

	if got := Ok(5).UnwrapOr(0); got != 5 {
		t.Fatalf("expected ok value 5, got %d", got)
	}
	errResult := Err[int](errors.New("boom"))
	if !errResult.IsErr() || errResult.UnwrapOr(9) != 9 {
		t.Fatalf("expected err result fallback")
	}
}

func TestApplyAndTagScanner(t *testing.T) {
	type options struct{ Enabled bool }
	cfg := options{}
	Apply(&cfg, func(o *options) { o.Enabled = true })
	if !cfg.Enabled {
		t.Fatalf("expected option applied")
	}

	type tagged struct {
		Count string `rfw:"signal"`
		Store string `rfw:"store:cart"`
		Ref   string `rfw:"ref"`
		Click string `rfw:"event:click:Save:prevent:stop"`
		Hist  string `rfw:"history:main:undo:redo"`
	}
	meta, err := TagScanner.Scan(tagged{})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(meta.Signals) != 1 || meta.Signals[0].Name != "Count" {
		t.Fatalf("unexpected signals: %+v", meta.Signals)
	}
	if len(meta.Stores) != 1 || meta.Stores[0].Name != "cart" {
		t.Fatalf("unexpected stores: %+v", meta.Stores)
	}
	if len(meta.Refs) != 1 || meta.Refs[0] != "Ref" {
		t.Fatalf("unexpected refs: %+v", meta.Refs)
	}
	if len(meta.Events) != 1 || meta.Events[0].DOMEvent != "click" || meta.Events[0].Handler != "Save" || len(meta.Events[0].Modifiers) != 2 {
		t.Fatalf("unexpected events: %+v", meta.Events)
	}
	if len(meta.Histories) != 1 || meta.Histories[0].Store != "main" {
		t.Fatalf("unexpected histories: %+v", meta.Histories)
	}
}
