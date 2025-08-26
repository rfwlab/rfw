//go:build js && wasm

package virtual

import (
	"fmt"
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/events"
)

// VirtualList renders only the portion of a list that is visible within its container.
type VirtualList struct {
	Container  jst.Value
	Total      int
	ItemHeight int
	Render     func(i int) string
	stopScroll func()
}

// NewVirtualList attaches a virtualized list to the element with the given id.
func NewVirtualList(containerID string, total, itemHeight int, render func(i int) string) *VirtualList {
	v := &VirtualList{
		Container:  dom.ByID(containerID),
		Total:      total,
		ItemHeight: itemHeight,
		Render:     render,
	}
	v.stopScroll = events.OnScroll(v.Container, func(jst.Value) {
		v.update()
	})
	v.update()
	return v
}

// update recalculates the visible range and mounts the required items.
func (v *VirtualList) update() {
	height := v.Container.Get("clientHeight").Int()
	scrollTop := v.Container.Get("scrollTop").Int()
	start := max(0, scrollTop/v.ItemHeight)
	end := min(v.Total, start+height/v.ItemHeight+1)

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
	if v.stopScroll != nil {
		v.stopScroll()
	}
}
