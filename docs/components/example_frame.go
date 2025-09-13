//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"

	core "github.com/rfwlab/rfw/v1/core"
	events "github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/example_frame.rtml
var exampleFrameTpl []byte

type exampleFrame struct {
	*core.HTMLComponent
}

func NewExampleFrame(props map[string]any) *core.HTMLComponent {
	ef := &exampleFrame{}
	ef.HTMLComponent = core.NewComponent("ExampleFrame", exampleFrameTpl, props).WithLifecycle(ef.mount, nil)
	return ef.HTMLComponent
}

func init() {
	core.MustRegisterComponent("ExampleFrame", func() core.Component { return NewExampleFrame(nil) })
}

func (e *exampleFrame) mount(hc *core.HTMLComponent) {
	wrapper := hc.GetRef("wrapper")
	if !wrapper.Truthy() {
		return
	}
	frame := hc.GetRef("frame")
	if uri, ok := hc.Props["uri"].(string); ok {
		uri = fmt.Sprintf(`%s?%s`, uri, hc.GetID())
		frame.SetAttr("src", uri)
	}
	codeTab := hc.GetRef("codeTab")
	previewTab := hc.GetRef("previewTab")
	codeDiv := hc.GetRef("codeDiv")
	codeEl := hc.GetRef("codeEl")
	filePathEl := hc.GetRef("filePath")
	previewDiv := hc.GetRef("previewDiv")
	if codeURL, ok := hc.Props["code"].(string); ok && codeURL != "" {
		filePathEl.SetText(codeURL)

		js.Fetch(codeURL).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			resp := args[0]
			return resp.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				code := args[0].String()
				codeEl.SetText(code)
				if h := js.Get("rfwHighlight"); h.Truthy() {
					res := h.Invoke(code, "go")
					if res.Truthy() && res.String() != "" {
						codeEl.SetHTML(res.String())
					} else if hljs := js.Get("hljs"); hljs.Truthy() {
						hljs.Call("highlightElement", codeEl.Value)
					}
				} else if hljs := js.Get("hljs"); hljs.Truthy() {
					hljs.Call("highlightElement", codeEl.Value)
				}
				return nil
			}))
		}))
	}
	codeCh := events.Listen("click", codeTab.Value)
	go func() {
		for range codeCh {
			codeDiv.RemoveClass("hidden")
			previewDiv.AddClass("hidden")
			codeTab.AddClass("border-b-2")
			codeTab.AddClass("border-red-500")
			codeTab.AddClass("text-red-500")
			previewTab.RemoveClass("border-b-2")
			previewTab.RemoveClass("border-red-500")
			previewTab.RemoveClass("text-red-500")
		}
	}()
	previewCh := events.Listen("click", previewTab.Value)
	go func() {
		for range previewCh {
			previewDiv.RemoveClass("hidden")
			codeDiv.AddClass("hidden")
			previewTab.AddClass("border-b-2")
			previewTab.AddClass("border-red-500")
			previewTab.AddClass("text-red-500")
			codeTab.RemoveClass("border-b-2")
			codeTab.RemoveClass("border-red-500")
			codeTab.RemoveClass("text-red-500")
		}
	}()
}
