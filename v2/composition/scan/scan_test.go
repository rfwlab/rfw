//go:build js && wasm

package scan

import (
	"reflect"
	"testing"

	"github.com/rfwlab/rfw/v2/state"
	types "github.com/rfwlab/rfw/v2/types"
)

type TestPage struct {
	Count    types.Int
	Name     types.String
	Active   types.Bool
	Price    types.Float
	Items    *types.Slice[string]
	Cart     *types.Store
	Input    *types.Ref
	Content  *types.View
	Visits   types.HInt
	Message  types.HString
	Logger   *types.Inject[any]
	Hist     *types.History
}

func TestScanDetectsSignalTypes(t *testing.T) {
	page := &TestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	signalNames := make(map[string]bool)
	for _, s := range meta.Signals {
		signalNames[s.Name] = true
	}
	// Count, Name, Active, Price, Visits, Message — all are signals
	// Visits and Message are also hosts, but they're still signals
	for _, name := range []string{"Count", "Name", "Active", "Price", "Visits", "Message"} {
		if !signalNames[name] {
			t.Errorf("expected signal %q, got signals: %v", name, meta.Signals)
		}
	}
	if len(meta.Signals) != 6 {
		t.Errorf("expected 6 signals, got %d: %v", len(meta.Signals), meta.Signals)
	}
}

func TestScanDetectsStore(t *testing.T) {
	page := &TestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(meta.Stores) != 1 || meta.Stores[0].Name != "Cart" {
		t.Errorf("expected store Cart, got: %v", meta.Stores)
	}
}

func TestScanDetectsRef(t *testing.T) {
	page := &TestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(meta.Refs) != 1 || meta.Refs[0].Name != "Input" {
		t.Errorf("expected ref Input, got: %v", meta.Refs)
	}
}

func TestScanDetectsView(t *testing.T) {
	page := &TestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(meta.Includes) != 1 || meta.Includes[0].Field != "Content" {
		t.Errorf("expected include Content, got: %v", meta.Includes)
	}
}

func TestScanDetectsHostTypes(t *testing.T) {
	page := &TestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	hostNames := make(map[string]bool)
	for _, h := range meta.Hosts {
		hostNames[h.Name] = true
	}
	for _, name := range []string{"Visits", "Message"} {
		if !hostNames[name] {
			t.Errorf("expected host %q, got hosts: %v", name, meta.Hosts)
		}
	}
	if len(meta.Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d: %v", len(meta.Hosts), meta.Hosts)
	}
}

func TestScanTemplateConvention(t *testing.T) {
	page := &TestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if meta.TemplateName != "TestPage" {
		t.Errorf("expected template name TestPage, got: %s", meta.TemplateName)
	}
}

func TestIsSignalTypeValue(t *testing.T) {
	var x types.Int
	typ := reflect.TypeOf(x)
	if !isSignalType(typ) {
		t.Errorf("expected types.Int (value type) to be detected as signal, got type: %v", typ)
	}
}

func TestIsSignalTypePointer(t *testing.T) {
	var x types.Int
	typ := reflect.TypeOf(&x)
	if !isSignalType(typ) {
		t.Errorf("expected *types.Int (pointer type) to be detected as signal, got type: %v", typ)
	}
}

func TestIsSignalTypeStateSignal(t *testing.T) {
	var x state.Signal[int]
	typ := reflect.TypeOf(x)
	if !isSignalType(typ) {
		t.Errorf("expected state.Signal[int] to be detected as signal, got type: %v", typ)
	}
}

func TestIsNotSignalType(t *testing.T) {
	typ := reflect.TypeOf("")
	if isSignalType(typ) {
		t.Error("expected string not to be detected as signal")
	}
}

func TestIsHostType(t *testing.T) {
	var x types.HInt
	typ := reflect.TypeOf(x)
	if !isHostType(typ) {
		t.Errorf("expected types.HInt to be detected as host type, got type: %v", typ)
	}
}

func TestIsHostTypePointer(t *testing.T) {
	var x types.HInt
	typ := reflect.TypeOf(&x)
	if !isHostType(typ) {
		t.Errorf("expected *types.HInt to be detected as host type, got type: %v", typ)
	}
}

type MethodTestPage struct {
	Count types.Int
}

func (m *MethodTestPage) Inc() { m.Count.Set(m.Count.Get() + 1) }
func (m *MethodTestPage) Dec() { m.Count.Set(m.Count.Get() - 1) }

func TestScanDiscoversMethods(t *testing.T) {
	page := &MethodTestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	handlerNames := make(map[string]bool)
	for _, ev := range meta.Events {
		handlerNames[ev.Handler] = true
	}
	if !handlerNames["Inc"] {
		t.Errorf("expected Inc handler, got events: %v", meta.Events)
	}
	if !handlerNames["Dec"] {
		t.Errorf("expected Dec handler, got events: %v", meta.Events)
	}
}

func TestScanDiscoversOnMount(t *testing.T) {
	// OnMount should NOT be in events, it should be handled separately
	page := &MethodTestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	for _, ev := range meta.Events {
		if ev.Handler == "OnMount" {
			t.Error("OnMount should not be in events list")
		}
	}
}

func TestScanDetectsInjection(t *testing.T) {
	page := &TestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(meta.Injections) != 1 || meta.Injections[0].Name != "Logger" {
		t.Errorf("expected injection Logger, got: %v", meta.Injections)
	}
}

func TestScanDetectsHistory(t *testing.T) {
	page := &TestPage{}
	meta, err := Scan(page)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	// History should be in Histories, not Stores
	if len(meta.Histories) != 1 || meta.Histories[0].Name != "Hist" {
		t.Errorf("expected history Hist, got: %v", meta.Histories)
	}
	// History should NOT appear in Stores
	for _, s := range meta.Stores {
		if s.Name == "Hist" {
			t.Error("History should not be in Stores slice")
		}
	}
}

func TestIsInjectType(t *testing.T) {
	var x types.Inject[any]
	typ := reflect.TypeOf(&x)
	if !isInjectType(typ) {
		t.Errorf("expected *Inject[any] to be detected as inject type, got type: %v", typ)
	}
	// Non-pointer should not match
	if isInjectType(reflect.TypeOf(x)) {
		t.Error("expected value Inject[any] to NOT be detected as inject type")
	}
}