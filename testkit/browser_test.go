//go:build js && wasm

package testkit

import (
	"testing"

	"github.com/rfwlab/rfw/v2/core"
)

func TestBrowserHarnessMountsAndQueries(t *testing.T) {
	component := core.NewHTMLComponent("Harness", []byte(`<root><button id="action">run</button></root>`), nil)
	component.SetComponent(component)
	component.Init(nil)

	harness := Mount(t, component)
	if harness.Query("#action").Text() != "run" {
		t.Fatalf("query returned wrong element: %s", harness.HTML())
	}
}

func TestBrowserHarnessAssertsLiveNodeState(t *testing.T) {
	component := core.NewHTMLComponent("HarnessLive", []byte(`<root><input id="field" value="initial"></root>`), nil)
	component.SetComponent(component)
	component.Init(nil)

	harness := Mount(t, component)
	field := harness.Query("#field")
	field.SetValue("typed")
	field.Call("focus")
	harness.AssertSameNode(t, "#field", field)
	harness.AssertActive(t, "#field")
	if got := harness.LiveValue("#field"); got != "typed" {
		t.Fatalf("live value = %q, want typed", got)
	}
}
