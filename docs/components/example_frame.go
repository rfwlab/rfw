//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
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
	root := dom.Doc().Query("[data-component-id='" + hc.GetID() + "']")
	if !root.Truthy() {
		return
	}
	wrapper := root.Get("firstElementChild")
	if code, ok := hc.Props["code"].(string); ok {
		wrapper.Get("dataset").Set("code", code)
	}
	iframe := wrapper.Call("querySelector", "iframe")
	if uri, ok := hc.Props["uri"].(string); ok {
		uri = fmt.Sprintf(`%s?%s`, uri, hc.GetID())
		iframe.Set("src", uri)
	}
	codeTab := wrapper.Call("querySelector", "#tab-code")
	previewTab := wrapper.Call("querySelector", "#tab-preview")
	codeDiv := wrapper.Call("querySelector", "#code")
	codeEl := codeDiv.Call("querySelector", "code")
	filePathEl := wrapper.Call("querySelector", "#file-path")
	previewDiv := wrapper.Call("querySelector", "#preview")
	codeURL := wrapper.Get("dataset").Get("code").String()
	if codeURL != "" {
		filePathEl.Set("textContent", codeURL)

		js.Fetch(codeURL).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			resp := args[0]
			return resp.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				code := args[0].String()
				codeEl.Set("textContent", code)
				if h := js.Get("rfwHighlight"); h.Truthy() {
					res := h.Invoke(code, "go")
					if res.Truthy() && res.String() != "" {
						codeEl.Set("innerHTML", res.String())
					} else if hljs := js.Get("hljs"); hljs.Truthy() {
						hljs.Call("highlightElement", codeEl)
					}
				} else if hljs := js.Get("hljs"); hljs.Truthy() {
					hljs.Call("highlightElement", codeEl)
				}
				return nil
			}))
		}))
	}
	codeCh := events.Listen("click", codeTab)
	go func() {
		for range codeCh {
			codeDiv.Get("classList").Call("remove", "hidden")
			previewDiv.Get("classList").Call("add", "hidden")
			codeTab.Get("classList").Call("add", "border-b-2", "border-red-500", "text-red-500")
			previewTab.Get("classList").Call("remove", "border-b-2", "border-red-500", "text-red-500")
		}
	}()
	previewCh := events.Listen("click", previewTab)
	go func() {
		for range previewCh {
			previewDiv.Get("classList").Call("remove", "hidden")
			codeDiv.Get("classList").Call("add", "hidden")
			previewTab.Get("classList").Call("add", "border-b-2", "border-red-500", "text-red-500")
			codeTab.Get("classList").Call("remove", "border-b-2", "border-red-500", "text-red-500")
		}
	}()
}
