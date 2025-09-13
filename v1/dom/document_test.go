//go:build js && wasm

package dom

import "testing"

func TestDocumentElementBasics(t *testing.T) {
	doc := Doc()
	el := doc.CreateElement("div")
	el.SetText("hello")
	if got := el.Text(); got != "hello" {
		t.Fatalf("Text() = %q", got)
	}
}
