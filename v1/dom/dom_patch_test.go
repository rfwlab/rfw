//go:build js && wasm

package dom

import (
	"testing"

	js "github.com/rfwlab/rfw/v1/js"
)

// Ensure UpdateDOM handles nodes without attributes (e.g. comments) without panicking.
func TestUpdateDOMSkipsNonElementNodes(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("div")
	root.Set("id", "root")
	body.Call("appendChild", root)
	defer root.Call("remove")

	SetInnerHTML(root, "<!--old-->")
	UpdateDOM("root", "<!--new-->")
}
