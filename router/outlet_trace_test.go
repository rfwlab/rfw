//go:build js && wasm

package router

import (
	"syscall/js"
	"testing"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/internal/rendertrace"
)

type routerTraceSnapshot struct {
	event       string
	componentID string
	parentID    string
	cause       string
	hasTimings  bool
}

func TestOutletTraceCoversRootMountAndRouteCommit(t *testing.T) {
	restoreTrace := rendertrace.SetEnabledForTest(true)
	defer restoreTrace()
	var records []routerTraceSnapshot
	callback := js.FuncOf(func(_ js.Value, args []js.Value) any {
		detail := args[0].Get("detail")
		cause := detail.Get("cause")
		record := routerTraceSnapshot{
			event:       detail.Get("event").String(),
			componentID: detail.Get("componentId").String(),
			hasTimings:  detail.Call("hasOwnProperty", "templateMs").Bool() && detail.Call("hasOwnProperty", "domMs").Bool(),
		}
		if value := detail.Get("parentComponentId"); value.Type() == js.TypeString {
			record.parentID = value.String()
		}
		if cause.Type() == js.TypeObject {
			record.cause = cause.Get("kind").String()
		}
		records = append(records, record)
		return nil
	})
	js.Global().Call("addEventListener", "rfw:render-trace", callback)
	defer func() {
		js.Global().Call("removeEventListener", "rfw:render-trace", callback)
		callback.Release()
	}()

	Reset()
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}
	outlet := NewOutlet()
	shell := core.NewHTMLComponent("TraceShell", []byte(`<root><aside>shell</aside>@include:outlet</root>`), nil)
	shell.SetComponent(shell)
	shell.AddDependency("outlet", outlet)
	shell.Init(nil)
	page := core.NewHTMLComponent("TracePage", []byte(`<root><main>page</main></root>`), nil)
	page.SetComponent(page)
	page.Init(nil)
	defer page.Unmount()
	defer shell.Unmount()

	Page("/trace-route", page)
	MountRoot(shell)
	Navigate("/trace-route")

	var shellCommit, pageCommit *routerTraceSnapshot
	for index := range records {
		record := &records[index]
		if record.event != "committed" {
			continue
		}
		switch record.componentID {
		case shell.ID:
			shellCommit = record
		case page.ID:
			pageCommit = record
		}
	}
	if shellCommit == nil || shellCommit.cause != "mount" || !shellCommit.hasTimings {
		t.Fatalf("root mount trace = %#v (all=%#v)", shellCommit, records)
	}
	if pageCommit == nil || pageCommit.cause != "router" || pageCommit.parentID != outlet.GetID() || !pageCommit.hasTimings {
		t.Fatalf("route commit trace = %#v (all=%#v)", pageCommit, records)
	}
}
