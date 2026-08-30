//go:build js && wasm

package hostclient

import (
	"strings"
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/core"
	dom "github.com/rfwlab/rfw/v2/dom"
	js "github.com/rfwlab/rfw/v2/js"
)

// The plain registrar keeps the signature applications pass to
// core.SetHostRegister as a function value.
var _ func(id, name string, vars []string) = RegisterComponent

// The owned registrar is what core.SetHostRegistrar takes.
var _ core.HostRegistrar = RegisterComponentOwned

func bindingSnapshot(name string) (componentBinding, bool) {
	mu.RLock()
	defer mu.RUnlock()
	binding, bound := bindings[name]
	return binding, bound
}

func pendingFor(name string) (inits, unsubscribes int) {
	mu.RLock()
	defer mu.RUnlock()
	for _, queued := range pending {
		if queued.name != name {
			continue
		}
		values, _ := queued.payload.(map[string]any)
		if values["init"] == true {
			inits++
		}
		if values["unsubscribe"] == true {
			unsubscribes++
		}
	}
	return inits, unsubscribes
}

// pendingLabels renders the queued messages for the given names in queue order,
// so a test can assert where a control sits and not only how many there are.
func pendingLabels(names ...string) []string {
	wanted := make(map[string]struct{}, len(names))
	for _, name := range names {
		wanted[name] = struct{}{}
	}
	mu.RLock()
	defer mu.RUnlock()
	labels := make([]string, 0, len(pending))
	for _, queued := range pending {
		if _, ok := wanted[queued.name]; !ok {
			continue
		}
		labels = append(labels, queued.name+"/"+pendingKind(queued.payload))
	}
	return labels
}

func pendingKind(payload any) string {
	values, ok := payload.(map[string]any)
	if !ok {
		return "message"
	}
	switch {
	case values["init"] == true:
		return "init"
	case values["unsubscribe"] == true:
		return "unsubscribe"
	}
	if cmd, ok := values["cmd"].(string); ok {
		return "cmd:" + cmd
	}
	return "message"
}

// forgetPending drops the messages a test queued, so package state does not
// leak into the next one.
func forgetPending(t *testing.T, name string) {
	t.Cleanup(func() {
		mu.Lock()
		filtered := pending[:0]
		for _, queued := range pending {
			if queued.name == name {
				continue
			}
			filtered = append(filtered, queued)
		}
		pending = filtered
		mu.Unlock()
	})
}

// setConnection makes the package treat c as its host connection, the way the
// connection loop does once a handshake completes. A nil c is the offline
// state, where registration controls go to the reconnect queue.
func setConnection(t *testing.T, c *hostConn) {
	t.Helper()
	mu.Lock()
	conn = c
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		conn = nil
		mu.Unlock()
	})
}

// setSnapshotBarrier installs the delivery barrier under the lock the read loop
// reads it with.
func setSnapshotBarrier(fn func(component string)) {
	mu.Lock()
	afterBindingSnapshot = fn
	mu.Unlock()
}

// A registration made through the plain registrar owns no cleanup: it lives
// until another registration replaces it, which is the behavior code calling
// RegisterComponent as a statement has always had.
func TestRegisterComponentBindsWithoutACleanup(t *testing.T) {
	name := "plain-binding"
	forgetPending(t, name)

	RegisterComponent("plain-root", name, []string{"value"})
	binding, bound := bindingSnapshot(name)
	if !bound || binding.id != "plain-root" {
		t.Fatalf("registration bound %#v bound=%v", binding, bound)
	}

	release := RegisterComponentOwned("owned-root", name, nil)
	if binding, _ := bindingSnapshot(name); binding.id != "owned-root" {
		t.Fatalf("the owned registration did not replace the plain one: %#v", binding)
	}
	release()
	if _, bound := bindingSnapshot(name); bound {
		t.Fatal("binding remained after the owned cleanup")
	}
}

func TestRegisterComponentCleanupReleasesTheBinding(t *testing.T) {
	name := "scoped-binding"
	setConnection(t, nil)
	forgetPending(t, name)

	release := RegisterComponentOwned("scoped-root", name, []string{"value"})
	if _, bound := bindingSnapshot(name); !bound {
		t.Fatal("registration did not bind the component")
	}
	if inits, _ := pendingFor(name); inits != 1 {
		t.Fatalf("queued init messages = %d, want 1", inits)
	}

	release()
	release()

	if _, bound := bindingSnapshot(name); bound {
		t.Fatal("binding remained after cleanup")
	}
	inits, unsubscribes := pendingFor(name)
	if inits != 0 {
		t.Fatalf("stale init messages = %d, want 0", inits)
	}
	if unsubscribes != 1 {
		t.Fatalf("queued unsubscribe messages = %d, want 1", unsubscribes)
	}
}

func TestStaleComponentCleanupKeepsTheReplacementBinding(t *testing.T) {
	name := "replacement-binding"
	forgetPending(t, name)

	first := RegisterComponentOwned("first-root", name, nil)
	second := RegisterComponentOwned("second-root", name, nil)
	first()

	binding, bound := bindingSnapshot(name)
	if !bound || binding.id != "second-root" {
		t.Fatalf("stale cleanup removed the replacement binding: %#v bound=%v", binding, bound)
	}
	second()
	if _, bound := bindingSnapshot(name); bound {
		t.Fatal("the replacement binding survived its own cleanup")
	}
}

// With a live connection the lifecycle talks to the host directly: one init on
// registration and one unsubscribe on cleanup, however often the cleanup runs,
// and nothing left queued for a reconnect.
func TestActiveConnectionCleanupSendsOneUnsubscribe(t *testing.T) {
	c, socket := dialFake(t)
	setConnection(t, c)
	name := "wired-ticker"
	forgetPending(t, name)

	release := RegisterComponentOwned("wired-root", name, nil)
	release()
	release()

	inits, unsubscribes := 0, 0
	for _, frame := range socket.sent() {
		if !strings.Contains(frame, `"component":"`+name+`"`) {
			continue
		}
		if strings.Contains(frame, `"init":true`) {
			inits++
		}
		if strings.Contains(frame, `"unsubscribe":true`) {
			unsubscribes++
		}
	}
	if inits != 1 || unsubscribes != 1 {
		t.Fatalf("frames sent: init=%d unsubscribe=%d, want 1 each", inits, unsubscribes)
	}
	if _, queued := pendingFor(name); queued != 0 {
		t.Fatalf("a live connection queued %d unsubscribes", queued)
	}
}

// Route entry and exit repeated: every mount owns one binding and one init, and
// leaves nothing behind when it unmounts. Offline, the queue keeps only the
// latest desired state for the name, so churn does not retain protocol work
// proportional to the number of cycles.
func TestComponentRegistrationCyclesRetainNothing(t *testing.T) {
	name := "cycled-binding"
	setConnection(t, nil)
	forgetPending(t, name)

	for cycle := 0; cycle < 100; cycle++ {
		release := RegisterComponentOwned("cycled-root", name, nil)
		if _, bound := bindingSnapshot(name); !bound {
			t.Fatalf("cycle %d: registration did not bind the component", cycle)
		}
		release()
		if _, bound := bindingSnapshot(name); bound {
			t.Fatalf("cycle %d: binding retained after cleanup", cycle)
		}
	}

	mu.RLock()
	_, tracked := bindingTokens[name]
	mu.RUnlock()
	if tracked {
		t.Fatal("registration bookkeeping retained the released name")
	}
	inits, unsubscribes := pendingFor(name)
	if inits != 0 || unsubscribes != 1 {
		t.Fatalf("queued messages: init=%d unsubscribe=%d, want 0 and 1", inits, unsubscribes)
	}
}

// The pending queue is a per-name desired state, not a log: a control is
// superseded where it stands, so the messages that are not registration
// controls, and the controls of other names, keep their exact place.
func TestPendingRegistrationControlKeepsOnlyTheLatestState(t *testing.T) {
	name := "queued-binding"
	other := "other-binding"
	setConnection(t, nil)
	forgetPending(t, name)
	forgetPending(t, other)

	assertQueue := func(step string, want ...string) {
		t.Helper()
		got := strings.Join(pendingLabels(name, other), " ")
		if expected := strings.Join(want, " "); got != expected {
			t.Fatalf("%s: queue = %q, want %q", step, got, expected)
		}
	}

	release := RegisterComponentOwned("queued-root", name, nil)
	Send(name, map[string]any{"cmd": "refresh"})
	otherRelease := RegisterComponentOwned("other-root", other, nil)
	assertQueue("after registration",
		name+"/init", name+"/cmd:refresh", other+"/init")

	release()
	assertQueue("after release",
		name+"/unsubscribe", name+"/cmd:refresh", other+"/init")

	second := RegisterComponentOwned("queued-root", name, nil)
	assertQueue("after re-registration",
		name+"/init", name+"/cmd:refresh", other+"/init")

	otherRelease()
	assertQueue("after the other release",
		name+"/init", name+"/cmd:refresh", other+"/unsubscribe")

	second()
	assertQueue("after the second release",
		name+"/unsubscribe", name+"/cmd:refresh", other+"/unsubscribe")
}

// A control decided with the connection up goes out on the wire, so the queue
// keeps neither it nor the state it superseded, and everything queued around it
// stays where it was.
func TestConnectedRegistrationControlDropsTheQueuedControl(t *testing.T) {
	name := "wired-queue-binding"
	other := "wired-queue-other"
	setConnection(t, nil)
	forgetPending(t, name)
	forgetPending(t, other)

	release := RegisterComponentOwned("wired-queue-root", name, nil)
	Send(name, map[string]any{"cmd": "refresh"})
	otherRelease := RegisterComponentOwned("wired-queue-other-root", other, nil)
	t.Cleanup(otherRelease)

	c, _ := dialFake(t)
	setConnection(t, c)
	release()

	got := strings.Join(pendingLabels(name, other), " ")
	want := strings.Join([]string{name + "/cmd:refresh", other + "/init"}, " ")
	if got != want {
		t.Fatalf("queue = %q, want %q", got, want)
	}
}

// This package installs the core registration hook from its own init, so a
// mounted component binds through the real boundary and unmounting releases it,
// with no test double in between.
func TestCoreComponentLifecycleOwnsItsBinding(t *testing.T) {
	setConnection(t, nil)
	forgetPending(t, "BoundaryHost")

	component := core.NewHTMLComponent("BoundaryComponent", []byte(`<root></root>`), nil)
	component.AddHostComponent("BoundaryHost")
	component.Init(nil)

	for cycle := 0; cycle < 100; cycle++ {
		component.Mount()
		component.Render()
		binding, bound := bindingSnapshot("BoundaryHost")
		if !bound || binding.id != component.ID {
			t.Fatalf("cycle %d: mount bound %#v, want the component id %q", cycle, binding, component.ID)
		}
		component.Unmount()
		if _, bound := bindingSnapshot("BoundaryHost"); bound {
			t.Fatalf("cycle %d: unmount left the binding in place", cycle)
		}
	}

	mu.RLock()
	_, tracked := bindingTokens["BoundaryHost"]
	mu.RUnlock()
	if tracked {
		t.Fatal("route churn retained the registration token")
	}
	inits, unsubscribes := pendingFor("BoundaryHost")
	if inits != 0 || unsubscribes != 1 {
		t.Fatalf("queued messages: init=%d unsubscribe=%d, want 0 and 1", inits, unsubscribes)
	}
}

// hostVarRoot builds a component root carrying one host variable element.
func hostVarRoot(t *testing.T, id string) dom.Element {
	t.Helper()
	root := dom.CreateElement("div")
	root.SetAttr("data-component-id", id)
	variable := dom.CreateElement("span")
	variable.SetAttr("data-host-var", "value")
	root.AppendChild(variable)
	dom.Doc().Body().AppendChild(root)
	t.Cleanup(func() { root.Call("remove") })
	return root
}

// appSentinel puts a host variable under #app, which is where
// dom.ComponentRoot falls back when a component id resolves to nothing. Host
// delivery must never take that fallback, so this element stays untouched.
func appSentinel(t *testing.T) dom.Element {
	t.Helper()
	app := dom.ByID("app")
	if app.IsNull() || app.IsUndefined() {
		app = dom.CreateElement("div")
		app.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(app)
		t.Cleanup(func() { app.Call("remove") })
	}
	sentinel := dom.CreateElement("span")
	sentinel.SetAttr("data-host-var", "value")
	sentinel.SetText("untouched")
	app.AppendChild(sentinel)
	t.Cleanup(func() { sentinel.Call("remove") })
	return sentinel
}

func hostVarText(root dom.Element) string {
	return root.Query(`[data-host-var="value"]`).Text()
}

func waitForHostVar(t *testing.T, root dom.Element, want string) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if hostVarText(root) == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("host variable = %q, want %q", hostVarText(root), want)
}

// A component that unmounts stops receiving host pushes: the frame that arrives
// after its cleanup reaches neither the root it owned, which route exit removed
// from the document, nor the #app shell the generic root lookup would fall back
// to, while a component registered afterwards keeps receiving its own.
func TestReleasedBindingIgnoresLateDelivery(t *testing.T) {
	conn, socket := dialFake(t)
	forgetPending(t, "released-ticker")
	forgetPending(t, "live-ticker")
	sentinel := appSentinel(t)
	releasedRoot := hostVarRoot(t, "released-root")
	liveRoot := hostVarRoot(t, "live-root")

	release := RegisterComponentOwned("released-root", "released-ticker", []string{"value"})
	t.Cleanup(release)
	go func() { _ = readOnce(conn, 2*time.Second) }()

	socket.deliver(`{"component":"released-ticker","payload":{"value":"1"},"sequence":1}`)
	waitForHostVar(t, releasedRoot, "1")

	// Route exit: the cleanup runs and the root it owned leaves the document.
	release()
	releasedRoot.Call("remove")
	socket.deliver(`{"component":"released-ticker","payload":{"value":"2"},"sequence":2}`)

	// Frames are applied in order, so a later frame reaching a live binding
	// proves the released one was already processed.
	releaseLive := RegisterComponentOwned("live-root", "live-ticker", []string{"value"})
	t.Cleanup(releaseLive)
	socket.deliver(`{"component":"live-ticker","payload":{"value":"3"},"sequence":3}`)
	waitForHostVar(t, liveRoot, "3")

	if got := hostVarText(releasedRoot); got != "1" {
		t.Fatalf("the released root was updated to %q", got)
	}
	if got := sentinel.Text(); got != "untouched" {
		t.Fatalf("a released frame fell back to #app and wrote %q", got)
	}
}

// withoutCSSEscape hides the CSS object for the duration of a test, which is
// the runtime where the root has to be found by scanning the attribute.
func withoutCSSEscape(t *testing.T) {
	t.Helper()
	original := js.Get("CSS")
	js.Set("CSS", js.Null())
	t.Cleanup(func() { js.Set("CSS", original) })
}

// RegisterComponent is public and takes any id, so an application can bind a
// root whose id is not a CSS identifier. Delivery has to resolve exactly that
// root: an id spliced into a selector raw would throw a DOMException out of
// querySelector, or select a root the binding does not own.
func TestDeliveryResolvesRootsWithCSSMetacharacters(t *testing.T) {
	cases := []struct {
		label     string
		component string
		id        string
		escape    bool
	}{
		{label: "css escape", component: "meta-escaped", id: `meta"root'\]:.#[ 1`, escape: true},
		{label: "attribute scan", component: "meta-scanned", id: `meta"root'\]:.#[ 2`, escape: false},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			if !tc.escape {
				withoutCSSEscape(t)
			}
			conn, socket := dialFake(t)
			forgetPending(t, tc.component)
			sentinel := appSentinel(t)
			decoy := hostVarRoot(t, "meta-decoy")
			root := hostVarRoot(t, tc.id)

			release := RegisterComponentOwned(tc.id, tc.component, []string{"value"})
			t.Cleanup(release)
			go func() { _ = readOnce(conn, 2*time.Second) }()

			socket.deliver(`{"component":"` + tc.component + `","payload":{"value":"1"},"sequence":1}`)
			waitForHostVar(t, root, "1")

			if got := hostVarText(decoy); got != "" {
				t.Fatalf("the frame also wrote %q into a root it does not own", got)
			}
			if got := sentinel.Text(); got != "untouched" {
				t.Fatalf("the frame fell back to #app and wrote %q", got)
			}
		})
	}
}

// The window a cleanup has to close is the one between the read loop
// snapshotting a binding and the delivery touching the DOM. The barrier runs
// inside exactly that window, on a frame the loop has already accepted, so the
// overlap is reproduced rather than raced for.
func TestCleanupDuringAnInFlightDeliveryWins(t *testing.T) {
	conn, socket := dialFake(t)
	forgetPending(t, "overlap-ticker")
	forgetPending(t, "overlap-live")
	sentinel := appSentinel(t)
	overlapRoot := hostVarRoot(t, "overlap-root")
	liveRoot := hostVarRoot(t, "overlap-live-root")

	release := RegisterComponentOwned("overlap-root", "overlap-ticker", []string{"value"})
	t.Cleanup(release)
	releaseLive := RegisterComponentOwned("overlap-live-root", "overlap-live", []string{"value"})
	t.Cleanup(releaseLive)

	barriers := make(chan struct{}, 1)
	setSnapshotBarrier(func(component string) {
		if component != "overlap-ticker" {
			return
		}
		setSnapshotBarrier(nil)
		release()
		overlapRoot.Call("remove")
		barriers <- struct{}{}
	})
	t.Cleanup(func() { setSnapshotBarrier(nil) })

	go func() { _ = readOnce(conn, 2*time.Second) }()
	socket.deliver(`{"component":"overlap-ticker","payload":{"value":"9"},"sequence":1}`)
	socket.deliver(`{"component":"overlap-live","payload":{"value":"10"},"sequence":2}`)
	waitForHostVar(t, liveRoot, "10")

	select {
	case <-barriers:
	default:
		t.Fatal("the delivery never reached the snapshot barrier")
	}
	if got := hostVarText(overlapRoot); got != "" {
		t.Fatalf("a delivery snapshotted before the cleanup still wrote %q", got)
	}
	if got := sentinel.Text(); got != "untouched" {
		t.Fatalf("the overlapping delivery fell back to #app and wrote %q", got)
	}
}
