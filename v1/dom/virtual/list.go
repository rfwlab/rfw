package virtual

import (
	"fmt"
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/dom"
)

// VirtualList renders only the portion of a list that is visible within its container.
type VirtualList struct {
	Container  jst.Value
	Total      int
	ItemHeight int
	Render     func(i int) string
	scrollFunc jst.Func
}

// NewVirtualList attaches a virtualized list to the element with the given id.
func NewVirtualList(containerID string, total, itemHeight int, render func(i int) string) *VirtualList {
	v := &VirtualList{
		Container:  dom.ByID(containerID),
		Total:      total,
		ItemHeight: itemHeight,
		Render:     render,
	}
	v.scrollFunc = jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		v.update()
		return nil
	})
	v.Container.Call("addEventListener", "scroll", v.scrollFunc)
	v.update()
	return v
}

// update recalculates the visible range and mounts the required items.
func (v *VirtualList) update() {
	height := v.Container.Get("clientHeight").Int()
	scrollTop := v.Container.Get("scrollTop").Int()
	start := scrollTop / v.ItemHeight
	if start < 0 {
		start = 0
	}
	end := start + height/v.ItemHeight + 1
	if end > v.Total {
		end = v.Total
	}

	offsetTop := start * v.ItemHeight
	html := fmt.Sprintf("<div style='height:%dpx'></div>", offsetTop)
	for i := start; i < end; i++ {
		html += v.Render(i)
	}
	offsetBottom := (v.Total - end) * v.ItemHeight
	html += fmt.Sprintf("<div style='height:%dpx'></div>", offsetBottom)
	dom.SetInnerHTML(v.Container, html)
}

// Destroy removes scroll listeners and cleans up resources.
func (v *VirtualList) Destroy() {
	v.Container.Call("removeEventListener", "scroll", v.scrollFunc)
	v.scrollFunc.Release()
}
