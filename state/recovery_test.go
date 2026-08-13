package state

import (
	"strings"
	"testing"
)

func captureCallbackPanics(t *testing.T) *[]string {
	t.Helper()
	previous := OnCallbackPanic
	contexts := []string{}
	OnCallbackPanic = func(_ any, context string, stack []byte) {
		if len(stack) == 0 {
			t.Error("recovered callback panic had no stack")
		}
		contexts = append(contexts, context)
	}
	t.Cleanup(func() { OnCallbackPanic = previous })
	return &contexts
}

func TestStoreContinuesNotificationsAfterListenerPanic(t *testing.T) {
	contexts := captureCallbackPanics(t)
	store := NewStore("recovery", WithModule("test"))
	called := 0
	store.OnChange("value", func(any) { panic("first listener") })
	store.OnChange("value", func(any) { called++ })

	store.Set("value", 1)
	store.Set("value", 2)

	if called != 2 {
		t.Fatalf("healthy listener calls = %d, want 2", called)
	}
	if len(*contexts) != 2 || !strings.Contains((*contexts)[0], "test.recovery.value") {
		t.Fatalf("unexpected recovery contexts: %v", *contexts)
	}
}

func TestSignalContinuesListenersAfterPanic(t *testing.T) {
	contexts := captureCallbackPanics(t)
	signal := NewSignal(0)
	called := 0
	signal.OnChange(func(int) { panic("first listener") })
	signal.OnChange(func(int) { called++ })

	signal.Set(1)
	signal.Set(2)

	if called != 2 {
		t.Fatalf("healthy listener calls = %d, want 2", called)
	}
	if len(*contexts) != 2 || (*contexts)[0] != "signal change listener" {
		t.Fatalf("unexpected recovery contexts: %v", *contexts)
	}
}

func TestEffectCanRunAgainAfterPanic(t *testing.T) {
	contexts := captureCallbackPanics(t)
	signal := NewSignal(0)
	runs := 0
	stop := Effect(func() func() {
		value := signal.Get()
		runs++
		if value == 1 {
			panic("effect update")
		}
		return nil
	})
	defer stop()

	signal.Set(1)
	signal.Set(2)

	if runs != 3 {
		t.Fatalf("effect runs = %d, want 3", runs)
	}
	if len(*contexts) != 1 || (*contexts)[0] != "signal effect" {
		t.Fatalf("unexpected recovery contexts: %v", *contexts)
	}
}

func TestRecoveryHookPanicDoesNotEscape(_ *testing.T) {
	previous := OnCallbackPanic
	OnCallbackPanic = func(any, string, []byte) { panic("broken reporter") }
	defer func() { OnCallbackPanic = previous }()

	store := NewStore("broken-reporter")
	store.OnChange("value", func(any) { panic("listener") })
	store.Set("value", 1)
}
