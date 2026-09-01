package state

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func waitResourceStatus[T any](t *testing.T, resource *Resource[T], status ResourceStatus) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for resource.Status() != status {
		if time.Now().After(deadline) {
			t.Fatalf("resource status = %q, want %q", resource.Status(), status)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestResourceLoadsAndMutates(t *testing.T) {
	resource := NewResource(func(context.Context) (int, error) { return 7, nil })
	defer resource.Close()
	waitResourceStatus(t, resource, ResourceReady)

	value, err := resource.Read()
	if err != nil || value != 7 {
		t.Fatalf("Read() = %d, %v", value, err)
	}
	resource.Mutate(9)
	value, err = resource.Read()
	if err != nil || value != 9 {
		t.Fatalf("Read() after Mutate = %d, %v", value, err)
	}
}

func TestResourceReportsErrors(t *testing.T) {
	want := errors.New("load failed")
	resource := NewResource(func(context.Context) (int, error) { return 0, want })
	defer resource.Close()
	waitResourceStatus(t, resource, ResourceError)

	if _, err := resource.Read(); !errors.Is(err, want) {
		t.Fatalf("Read() error = %v", err)
	}
}

func TestKeyedResourcesDeduplicateLoads(t *testing.T) {
	const key = "resource-deduplicate"
	ClearResourceCache(key)
	t.Cleanup(func() { ClearResourceCache(key) })
	start := make(chan struct{})
	var calls atomic.Int32
	loader := func(context.Context) (int, error) {
		calls.Add(1)
		<-start
		return 11, nil
	}
	first := NewResource(loader, WithResourceKey(key))
	second := NewResource(loader, WithResourceKey(key))
	defer first.Close()
	defer second.Close()
	close(start)
	waitResourceStatus(t, first, ResourceReady)
	waitResourceStatus(t, second, ResourceReady)

	if calls.Load() != 1 {
		t.Fatalf("loader calls = %d", calls.Load())
	}
}

func TestResourceCloseCancelsLoad(t *testing.T) {
	cancelled := make(chan struct{})
	resource := NewResource(func(ctx context.Context) (int, error) {
		<-ctx.Done()
		close(cancelled)
		return 0, ctx.Err()
	})
	resource.Close()

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("loader context was not cancelled")
	}
}

func TestResourceLoaderPanicBecomesError(t *testing.T) {
	resource := NewResource(func(context.Context) (int, error) {
		panic("loader")
	})
	defer resource.Close()

	waitResourceStatus(t, resource, ResourceError)
	if resource.Error() == nil {
		t.Fatal("resource panic did not become an error")
	}
}
