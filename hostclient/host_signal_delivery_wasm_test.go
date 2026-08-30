//go:build js && wasm

package hostclient

import (
	"sync"
	"testing"
	"time"

	dom "github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

// setHostSignalWriteHook installs the delivery hook under the lock the read
// loop reads it with.
func setHostSignalWriteHook(fn func(component, name string)) {
	mu.Lock()
	beforeHostSignalWrite = fn
	mu.Unlock()
}

// hostSignalRoot builds the root a binding owns and registers one host signal
// per name against it, which is the state a rendered component leaves behind.
func hostSignalRoot(t *testing.T, id string, names ...string) map[string]*state.Signal[string] {
	t.Helper()
	root := dom.CreateElement("div")
	root.SetAttr("data-component-id", id)
	dom.Doc().Body().AppendChild(root)
	signals := make(map[string]*state.Signal[string], len(names))
	for _, name := range names {
		signal := state.NewSignal("initial")
		dom.RegisterSignal(id, name, signal)
		signals[name] = signal
	}
	t.Cleanup(func() {
		dom.RemoveComponentSignals(id)
		root.Call("remove")
	})
	return signals
}

func waitForSignal(t *testing.T, signal *state.Signal[string], want string) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if signal.Get() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("host signal = %q, want %q", signal.Get(), want)
}

// untouched reports the signals of a component that still hold their initial
// value, so a test can count how much of a frame landed.
func untouched(signals map[string]*state.Signal[string]) int {
	count := 0
	for _, signal := range signals {
		if signal.Get() == "initial" {
			count++
		}
	}
	return count
}

// The window a cleanup has to close for host signals is the one between the
// ownership check a delivery makes and the write it was about to perform. The
// hook runs inside exactly that window, on a frame already accepted, so the
// overlap is reproduced rather than raced for: the cleanup returns first, and
// the frame it overlapped leaves the released component's signals alone.
func TestCleanupBetweenTheSignalCheckAndTheWriteWins(t *testing.T) {
	conn, socket := dialFake(t)
	forgetPending(t, "signal-ticker")
	forgetPending(t, "signal-live")
	released := hostSignalRoot(t, "signal-root", "value")
	live := hostSignalRoot(t, "signal-live-root", "value")

	release := RegisterComponentOwned("signal-root", "signal-ticker", []string{"value"})
	t.Cleanup(release)
	releaseLive := RegisterComponentOwned("signal-live-root", "signal-live", []string{"value"})
	t.Cleanup(releaseLive)

	cleanupReturned := make(chan struct{})
	setHostSignalWriteHook(func(component, name string) {
		if component != "signal-ticker" {
			return
		}
		setHostSignalWriteHook(nil)
		// Route exit runs on its own goroutine, as it does in a browser, and
		// the delivery waits for the cleanup to return before its write.
		go func() {
			release()
			close(cleanupReturned)
		}()
		select {
		case <-cleanupReturned:
		case <-time.After(2 * time.Second):
			t.Error("the cleanup did not return while a delivery was mid-frame")
		}
	})
	t.Cleanup(func() { setHostSignalWriteHook(nil) })

	go func() { _ = readOnce(conn, 2*time.Second) }()
	socket.deliver(`{"component":"signal-ticker","payload":{"value":"stale"},"sequence":1}`)
	// Frames are applied in order, so a later frame reaching a live binding
	// proves the overlapped one was already processed.
	socket.deliver(`{"component":"signal-live","payload":{"value":"fresh"},"sequence":2}`)
	waitForSignal(t, live["value"], "fresh")

	select {
	case <-cleanupReturned:
	default:
		t.Fatal("the delivery never reached the signal write hook")
	}
	if got := released["value"].Get(); got != "initial" {
		t.Fatalf("a frame that overlapped the cleanup wrote %q into a released host signal", got)
	}
}

// A host signal setter that unmounts its own component re-enters the release
// from inside the frame being applied. The release must not wait on the
// delivery it is running under, and the rest of that frame belongs to the
// registration the setter just dropped.
func TestSetterReleasingItsOwnBindingEndsTheFrame(t *testing.T) {
	conn, socket := dialFake(t)
	forgetPending(t, "reentrant-ticker")
	forgetPending(t, "reentrant-live")
	signals := hostSignalRoot(t, "reentrant-root", "first", "second")
	live := hostSignalRoot(t, "reentrant-live-root", "value")

	release := RegisterComponentOwned("reentrant-root", "reentrant-ticker", []string{"first", "second"})
	t.Cleanup(release)
	releaseLive := RegisterComponentOwned("reentrant-live-root", "reentrant-live", []string{"value"})
	t.Cleanup(releaseLive)

	var unmount sync.Once
	for _, signal := range signals {
		signal.OnChange(func(string) { unmount.Do(release) })
	}

	go func() { _ = readOnce(conn, 2*time.Second) }()
	socket.deliver(`{"component":"reentrant-ticker","payload":{"first":"one","second":"two"},"sequence":1}`)
	// The witness frame is what proves the read loop came back from the
	// re-entrant release instead of deadlocking inside it.
	socket.deliver(`{"component":"reentrant-live","payload":{"value":"fresh"},"sequence":2}`)
	waitForSignal(t, live["value"], "fresh")

	if remaining := untouched(signals); remaining != 1 {
		t.Fatalf("host signals left untouched = %d, want 1: the release from a setter did not end the frame", remaining)
	}
	if _, bound := bindingSnapshot("reentrant-ticker"); bound {
		t.Fatal("the binding survived the release its own setter ran")
	}
}

// A setter that mounts the component again re-enters registration from inside
// the frame. The replacement is what stays bound, and the frame the superseded
// registration was carrying ends where it was superseded.
func TestSetterRebindingItsComponentEndsTheFrame(t *testing.T) {
	conn, socket := dialFake(t)
	forgetPending(t, "rebound-ticker")
	forgetPending(t, "rebound-live")
	signals := hostSignalRoot(t, "rebound-root", "first", "second")
	live := hostSignalRoot(t, "rebound-live-root", "value")

	release := RegisterComponentOwned("rebound-root", "rebound-ticker", []string{"first", "second"})
	t.Cleanup(release)
	releaseLive := RegisterComponentOwned("rebound-live-root", "rebound-live", []string{"value"})
	t.Cleanup(releaseLive)

	var remount sync.Once
	rebound := make(chan func(), 1)
	for _, signal := range signals {
		signal.OnChange(func(string) {
			remount.Do(func() {
				rebound <- RegisterComponentOwned("remounted-root", "rebound-ticker", []string{"first"})
			})
		})
	}

	go func() { _ = readOnce(conn, 2*time.Second) }()
	socket.deliver(`{"component":"rebound-ticker","payload":{"first":"one","second":"two"},"sequence":1}`)
	socket.deliver(`{"component":"rebound-live","payload":{"value":"fresh"},"sequence":2}`)
	waitForSignal(t, live["value"], "fresh")

	select {
	case releaseRemounted := <-rebound:
		t.Cleanup(releaseRemounted)
	default:
		t.Fatal("the setter never re-registered the component")
	}
	if remaining := untouched(signals); remaining != 1 {
		t.Fatalf("host signals left untouched = %d, want 1: the registration from a setter did not end the frame", remaining)
	}
	binding, bound := bindingSnapshot("rebound-ticker")
	if !bound || binding.id != "remounted-root" {
		t.Fatalf("the replacement registration is not the live one: %#v bound=%v", binding, bound)
	}
}
