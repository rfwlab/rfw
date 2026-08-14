//go:build js && wasm

package dom

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkPatchKeyedList(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			root := CreateElement("root")
			componentID := "benchmark-keyed-" + strconv.Itoa(size)
			root.SetAttr("data-component-id", componentID)
			forward := benchmarkListHTML(componentID, size, false)
			reverse := benchmarkListHTML(componentID, size, true)
			root.SetHTML(strings.TrimSuffix(strings.TrimPrefix(forward, `<root data-component-id="`+componentID+`">`), "</root>"))
			recordRenderedTree(root.Value)
			Doc().Body().AppendChild(root)
			b.Cleanup(func() { root.Call("remove") })

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%2 == 0 {
					patchInnerHTML(root.Value, reverse)
				} else {
					patchInnerHTML(root.Value, forward)
				}
			}
		})
	}
}

func benchmarkListHTML(componentID string, size int, reverse bool) string {
	var html strings.Builder
	fmt.Fprintf(&html, `<root data-component-id="%s"><ul>`, componentID)
	for position := 0; position < size; position++ {
		item := position
		if reverse {
			item = size - position - 1
		}
		fmt.Fprintf(&html, `<li data-for="benchmark" data-key="%d">row-%d</li>`, item, item)
	}
	html.WriteString(`</ul></root>`)
	return html.String()
}
