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

func TestDocumentHead(t *testing.T) {
	doc := Doc()
	if node := doc.Head().Get("nodeName").String(); node != "HEAD" {
		t.Fatalf("Head() node = %q", node)
	}
}
