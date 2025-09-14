//go:build js && wasm

package components

import (
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/rfwlab/rfw/v1/composition"
	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/http"
	highlight "github.com/rfwlab/rfw/v1/plugins/highlight"
)

//go:embed templates/example_frame.rtml
var exampleFrameTpl []byte

func NewExampleFrame(props map[string]any) *core.HTMLComponent {
	cmp := composition.Wrap(core.NewComponent("ExampleFrame", exampleFrameTpl, props))

	cmp.SetOnMount(func(*core.HTMLComponent) {
		wrapper := cmp.GetRef("wrapper")
		if !wrapper.Truthy() {
			return
		}
		frame := cmp.GetRef("frame")
		if uri, ok := cmp.Props["uri"].(string); ok {
			uri = fmt.Sprintf(`%s?%s`, uri, cmp.GetID())
			frame.SetAttr("src", uri)
		}
		codeTab := cmp.GetRef("codeTab")
		previewTab := cmp.GetRef("previewTab")
		codeDiv := cmp.GetRef("codeDiv")
		codeEl := cmp.GetRef("codeEl")
		filePathEl := cmp.GetRef("filePath")
		previewDiv := cmp.GetRef("previewDiv")
		if codeURL, ok := cmp.Props["code"].(string); ok && codeURL != "" {
			filePathEl.SetText(codeURL)
			go func() {
				for {
					code, err := http.FetchText(codeURL)
					if err != nil {
						if errors.Is(err, http.ErrPending) {
							time.Sleep(50 * time.Millisecond)
							continue
						}
						filePathEl.SetText(err.Error())
						return
					}
					codeEl.SetText(code)
					highlight.HighlightAll()
					return
				}
			}()
		}

		codeTab.OnClick(func(dom.Event) {
			codeDiv.RemoveClass("hidden")
			previewDiv.AddClass("hidden")
			codeTab.AddClass("border-b-2")
			codeTab.AddClass("border-red-500")
			codeTab.AddClass("text-red-500")
			previewTab.RemoveClass("border-b-2")
			previewTab.RemoveClass("border-red-500")
			previewTab.RemoveClass("text-red-500")
		})

		previewTab.OnClick(func(dom.Event) {
			previewDiv.RemoveClass("hidden")
			codeDiv.AddClass("hidden")
			previewTab.AddClass("border-b-2")
			previewTab.AddClass("border-red-500")
			previewTab.AddClass("text-red-500")
			codeTab.RemoveClass("border-b-2")
			codeTab.RemoveClass("border-red-500")
			codeTab.RemoveClass("text-red-500")
		})
	})

	return cmp.Unwrap()
}

func init() {
	core.MustRegisterComponent("ExampleFrame", func() core.Component { return NewExampleFrame(nil) })
}
