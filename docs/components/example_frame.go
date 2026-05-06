//go:build js && wasm

package components

import (
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/http"
	highlight "github.com/rfwlab/rfw/v2/plugins/highlight"
	"github.com/rfwlab/rfw/v2/types"
)

//go:embed templates/example_frame.rtml
var exampleFrameTpl []byte

type ExampleFrame struct {
	*core.HTMLComponent
	Wrapper   *types.Ref
	Frame     *types.Ref
	CodeTab   *types.Ref
	PreviewTab *types.Ref
	CodeDiv   *types.Ref
	CodeEl    *types.Ref
	FilePath  *types.Ref
	PreviewDiv *types.Ref
}

func NewExampleFrame(props map[string]any) *core.HTMLComponent {
	c := &ExampleFrame{}
	c.HTMLComponent = core.NewComponent("ExampleFrame", exampleFrameTpl, props).WithLifecycle(c.mount, nil)
	c.SetComponent(c)
	return c.HTMLComponent
}

func (c *ExampleFrame) mount(hc *core.HTMLComponent) {
	wrapper := c.Wrapper.Get()
	if !wrapper.Truthy() {
		return
	}

	if uri, ok := c.HTMLComponent.Props["uri"].(string); ok {
		uri = fmt.Sprintf(`%s?%s`, uri, c.HTMLComponent.GetID())
		c.Frame.Get().Call("setAttribute", "src", uri)
	}

	if codeURL, ok := c.HTMLComponent.Props["code"].(string); ok && codeURL != "" {
		filePathEl := dom.Element{Value: c.FilePath.Get()}
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
				dom.Element{Value: c.CodeEl.Get()}.SetText(code)
				highlight.HighlightAll()
				return
			}
		}()
	}

	codeTab := dom.Element{Value: c.CodeTab.Get()}
	previewTab := dom.Element{Value: c.PreviewTab.Get()}
	codeDiv := dom.Element{Value: c.CodeDiv.Get()}
	previewDiv := dom.Element{Value: c.PreviewDiv.Get()}

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
}

func init() {
	core.MustRegisterComponent("ExampleFrame", func() core.Component { return NewExampleFrame(nil) })
}