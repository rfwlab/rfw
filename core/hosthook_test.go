//go:build js && wasm

package core

import "testing"

// hostCall is one registration the hook observed.
type hostCall struct {
	id, name string
	vars     []string
}

// recordHostCalls installs the legacy hook and collects what it receives.
func recordHostCalls(t *testing.T, calls *[]hostCall) {
	t.Helper()
	prev := hostRegisterComponent
	t.Cleanup(func() { hostRegisterComponent = prev })
	SetHostRegister(func(id, name string, vars []string) {
		*calls = append(*calls, hostCall{id: id, name: name, vars: vars})
	})
}

// Mounting a component that declares host components goes through the hook the
// hostclient package installs, once per declared name and in declaration order.
func TestHostRegisterHook(t *testing.T) {
	var calls []hostCall
	recordHostCalls(t, &calls)

	c := NewHTMLComponent("HookHost", []byte(`<root></root>`), nil)
	c.AddHostComponent("Counter")
	c.AddHostComponent("Clock")
	c.Init(nil)
	c.Render()
	c.Mount()

	if len(calls) != 2 {
		t.Fatalf("expected 2 host registrations, got %d: %+v", len(calls), calls)
	}
	if calls[0].name != "Counter" || calls[1].name != "Clock" {
		t.Fatalf("unexpected registration order: %+v", calls)
	}
	if calls[0].id != c.ID {
		t.Fatalf("expected component id %q, got %q", c.ID, calls[0].id)
	}
}

// Ownership is mounted-bound: a component that is only rendered registers
// nothing, because no Unmount would ever come to release it.
func TestHostRegisterSkipsRenderOnlyComponent(t *testing.T) {
	var calls []hostCall
	recordHostCalls(t, &calls)

	c := NewHTMLComponent("RenderOnlyHost", []byte(`<root>{h:value}</root>`), nil)
	c.AddHostComponent("Counter")
	c.Init(nil)
	for i := 0; i < 5; i++ {
		c.RenderFresh()
	}
	c.Unmount()

	if len(calls) != 0 {
		t.Fatalf("a render-only component registered %d host bindings: %+v", len(calls), calls)
	}
}

// Render before mount is the framework's own order: the render discovers the
// host variables and the mount binds once with them.
func TestHostRegisterAfterRenderThenMount(t *testing.T) {
	var calls []hostCall
	recordHostCalls(t, &calls)

	c := NewHTMLComponent("RenderThenMountHost", []byte(`<root>{h:value}</root>`), nil)
	c.AddHostComponent("Counter")
	c.Init(nil)
	c.Render()
	c.Mount()
	c.Render()

	if len(calls) != 1 {
		t.Fatalf("host registrations = %d, want 1: %+v", len(calls), calls)
	}
	if len(calls[0].vars) != 1 || calls[0].vars[0] != "value" {
		t.Fatalf("registered vars = %v, want the ones the render discovered", calls[0].vars)
	}
}

// Mount before render binds at mount time, with whatever the template has
// discovered so far, and the render that follows adds no second registration.
func TestHostRegisterAfterMountThenRender(t *testing.T) {
	var calls []hostCall
	recordHostCalls(t, &calls)

	c := NewHTMLComponent("MountThenRenderHost", []byte(`<root>{h:value}</root>`), nil)
	c.AddHostComponent("Counter")
	c.Init(nil)
	c.Mount()
	if len(calls) != 1 {
		t.Fatalf("mount registrations = %d, want 1: %+v", len(calls), calls)
	}
	if len(calls[0].vars) != 0 {
		t.Fatalf("registered vars = %v, want none discovered yet", calls[0].vars)
	}
	c.Render()
	c.RenderFresh()
	if len(calls) != 1 {
		t.Fatalf("host registrations = %d after rendering, want 1: %+v", len(calls), calls)
	}
}

// A nil hook is the client-only case: the lifecycle must not panic.
func TestHostRegisterHookNil(t *testing.T) {
	prev := hostRegisterComponent
	t.Cleanup(func() { hostRegisterComponent = prev })
	hostRegisterComponent = nil

	c := NewHTMLComponent("NoHook", []byte(`<root></root>`), nil)
	c.AddHostComponent("Counter")
	c.Init(nil)
	if got := c.Render(); got == "" {
		t.Fatal("expected a rendered template with no host hook installed")
	}
	c.Mount()
	c.Unmount()
}

// The legacy hook returns no cleanup. Unmounting a component that used it must
// still complete instead of calling a nil release.
func TestLegacyHostRegisterUnmountsWithoutCleanup(t *testing.T) {
	prev := hostRegisterComponent
	t.Cleanup(func() { hostRegisterComponent = prev })
	registrations := 0
	SetHostRegister(func(string, string, []string) { registrations++ })

	c := NewHTMLComponent("LegacyHook", []byte(`<root></root>`), nil)
	c.AddHostComponent("Counter")
	c.Init(nil)
	c.Mount()
	c.Render()
	c.Unmount()

	if registrations != 1 {
		t.Fatalf("host registrations = %d, want 1", registrations)
	}
}

// A host component is bound once per mounted lifecycle: repeated renders reuse
// the live binding instead of re-initializing the SSC feed.
func TestHostRegistrarBindsOncePerMountedLifecycle(t *testing.T) {
	prev := hostRegisterComponent
	t.Cleanup(func() { hostRegisterComponent = prev })
	registrations := 0
	releases := 0
	SetHostRegistrar(func(string, string, []string) func() {
		registrations++
		return func() { releases++ }
	})

	c := NewHTMLComponent("RenderedHost", []byte(`<root></root>`), nil)
	c.AddHostComponent("Counter")
	c.Init(nil)
	c.Mount()
	for i := 0; i < 5; i++ {
		c.RenderFresh()
	}

	if registrations != 1 || releases != 0 {
		t.Fatalf("registrations = %d, releases = %d, want 1 and 0", registrations, releases)
	}
}

// A remount whose render comes straight from the template cache still owns its
// bindings: the markup is identical, the previous registration is not.
func TestHostRegistrarRebindsOnCachedRemount(t *testing.T) {
	prev := hostRegisterComponent
	t.Cleanup(func() { hostRegisterComponent = prev })
	registrations := 0
	releases := 0
	SetHostRegistrar(func(string, string, []string) func() {
		registrations++
		return func() { releases++ }
	})

	c := NewHTMLComponent("CachedHost", []byte(`<root>{h:value}</root>`), nil)
	c.AddHostComponent("Counter")
	c.Init(nil)
	c.Mount()
	first := c.Render()
	c.Unmount()

	if registrations != 1 || releases != 1 {
		t.Fatalf("first cycle: registrations = %d, releases = %d, want 1 each", registrations, releases)
	}

	c.Mount()
	if second := c.Render(); second != first {
		t.Fatalf("the remount re-rendered instead of using the cache:\n%q\n%q", second, first)
	}
	if registrations != 2 || releases != 1 {
		t.Fatalf("cached remount: registrations = %d, releases = %d, want 2 and 1", registrations, releases)
	}
	c.Unmount()
	if releases != 2 {
		t.Fatalf("releases = %d after the second unmount, want 2", releases)
	}
}

// Route exit must not retain a host binding. Every mount owns its own
// registration and releases exactly that one on unmount, including when the
// render comes from the template cache.
func TestHostRegistrarReleasedOnEveryUnmount(t *testing.T) {
	prev := hostRegisterComponent
	t.Cleanup(func() { hostRegisterComponent = prev })
	active := map[string]int{}
	registrations := 0
	releases := 0
	SetHostRegistrar(func(_, name string, _ []string) func() {
		registrations++
		active[name]++
		return func() {
			releases++
			active[name]--
		}
	})

	c := NewHTMLComponent("CycledHost", []byte(`<root></root>`), nil)
	c.AddHostComponent("Counter")
	c.Init(nil)
	for cycle := 0; cycle < 100; cycle++ {
		c.Mount()
		c.Render()
		if active["Counter"] != 1 {
			t.Fatalf("cycle %d: active bindings = %d, want 1", cycle, active["Counter"])
		}
		c.Unmount()
		if active["Counter"] != 0 {
			t.Fatalf("cycle %d: binding retained after unmount", cycle)
		}
	}
	if registrations != 100 || releases != 100 {
		t.Fatalf("registrations = %d, releases = %d, want 100 each", registrations, releases)
	}
}
