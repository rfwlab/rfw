//go:build js && wasm

package components

import (
	_ "embed"
	"errors"
	"fmt"
	"time"

	core "github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/http"
	highlight "github.com/rfwlab/rfw/v2/plugins/highlight"
	"github.com/rfwlab/rfw/v2/types"
)

//go:embed templates/example_frame.rtml
var exampleFrameTpl []byte

type ExampleFrame struct {
	hc         *core.HTMLComponent
	Wrapper    *types.Ref
	Frame      *types.Ref
	CodeTab    *types.Ref
	PreviewTab *types.Ref
	CodeDiv    *types.Ref
	CodeEl     *types.Ref
	FilePath   *types.Ref
	PreviewDiv *types.Ref
}

func NewExampleFrame(props map[string]any) *core.HTMLComponent {
	f := &ExampleFrame{}
	c := core.NewComponentWith("ExampleFrame", exampleFrameTpl, props, f)
	f.hc = c
	c.Init(nil)
	return c
}

func (f *ExampleFrame) OnMount() {
	if !f.Wrapper.Get().Truthy() {
		return
	}

	if uri, ok := f.hc.Props["uri"].(string); ok {
		uri = fmt.Sprintf(`%s?%s`, uri, f.hc.GetID())
		f.Frame.Get().Call("setAttribute", "src", uri)
	}

	if codeURL, ok := f.hc.Props["code"].(string); ok && codeURL != "" {
		filePathEl := dom.Element{Value: f.FilePath.Get()}
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
				dom.Element{Value: f.CodeEl.Get()}.SetText(code)
				highlight.HighlightAll()
				return
			}
		}()
	}

	codeTab := dom.Element{Value: f.CodeTab.Get()}
	previewTab := dom.Element{Value: f.PreviewTab.Get()}
	codeDiv := dom.Element{Value: f.CodeDiv.Get()}
	previewDiv := dom.Element{Value: f.PreviewDiv.Get()}

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

func (f *ExampleFrame) Render() string          { return f.hc.Render() }
func (f *ExampleFrame) Mount()                  { f.hc.Mount() }
func (f *ExampleFrame) Unmount()                { f.hc.Unmount() }
func (f *ExampleFrame) OnUnmount()              {}
func (f *ExampleFrame) GetName() string         { return f.hc.GetName() }
func (f *ExampleFrame) GetID() string           { return f.hc.GetID() }
func (f *ExampleFrame) SetSlots(s map[string]any) { f.hc.SetSlots(s) }

func (f *ExampleFrame) IsMounted() bool {
	return f.hc.IsMounted()
}

func (f *ExampleFrame) OnParams(map[string]string) {}

func init() {
	core.MustRegisterComponent("ExampleFrame", func() core.Component { return NewExampleFrame(nil) })
}