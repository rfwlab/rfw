//go:build js && wasm

package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

func TestSuspenseMountState(t *testing.T) {
	s := NewSuspense(func() (string, error) { return "ready", nil }, "loading")
	if s.IsMounted() {
		t.Fatal("new Suspense should not be mounted")
	}
	s.Mount()
	if !s.IsMounted() {
		t.Fatal("Suspense should be mounted after Mount")
	}
	s.Unmount()
	if s.IsMounted() {
		t.Fatal("Suspense should not be mounted after Unmount")
	}
}

func TestSuspenseUpdatesWhenResourceResolves(t *testing.T) {
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	release := make(chan struct{})
	resource := state.NewResource(func(context.Context) (string, error) {
		<-release
		return "ready", nil
	})
	defer resource.Close()

	suspense := NewSuspense(func() (string, error) {
		value, err := resource.Read()
		return "<p>" + value + "</p>", err
	}, "<p>loading</p>")
	dom.UpdateDOM(suspense.GetID(), suspense.Render())
	suspense.Mount()
	defer suspense.Unmount()

	if html := dom.ComponentRoot(suspense.GetID()).HTML(); !strings.Contains(html, "loading") {
		t.Fatalf("fallback missing: %s", html)
	}
	close(release)

	deadline := time.Now().Add(time.Second)
	for {
		if html := dom.ComponentRoot(suspense.GetID()).HTML(); strings.Contains(html, "ready") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("resolved content missing: %s", dom.ComponentRoot(suspense.GetID()).HTML())
		}
		time.Sleep(time.Millisecond)
	}
}

func TestSuspenseCoalescesReactiveDOMUpdates(t *testing.T) {
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}

	value := state.NewSignal("zero")
	suspense := NewSuspense(func() (string, error) {
		return "<p>" + value.Get() + "</p>", nil
	}, "<p>loading</p>")
	dom.UpdateDOM(suspense.GetID(), suspense.Render())
	suspense.Mount()
	defer suspense.Unmount()

	previousHook := dom.TemplateHook
	commits := 0
	dom.TemplateHook = func(componentID, _ string) {
		if componentID == suspense.GetID() {
			commits++
		}
	}
	defer func() { dom.TemplateHook = previousHook }()

	for i := 0; i < 10; i++ {
		value.Set(fmt.Sprintf("value-%d", i))
	}
	waitForRenderFlush()

	if commits != 1 {
		t.Fatalf("Suspense update burst produced %d DOM commits, want 1", commits)
	}
	if html := dom.ComponentRoot(suspense.GetID()).HTML(); !strings.Contains(html, "value-9") {
		t.Fatalf("Suspense committed stale content: %s", html)
	}
}
