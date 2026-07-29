//go:build js && wasm

package http

import (
	"testing"

	"github.com/rfwlab/rfw/v2/js"
)

func TestRequestOptionsApply(t *testing.T) {
	init := RequestOptions{
		Method:  "POST",
		Headers: map[string]string{"X-Test": "1"},
		Body:    `{"a":1}`,
	}.apply()

	if got := init.Get("method").String(); got != "POST" {
		t.Fatalf("method = %q", got)
	}
	if got := init.Get("headers").Get("X-Test").String(); got != "1" {
		t.Fatalf("header = %q", got)
	}
	if got := init.Get("body").String(); got != `{"a":1}` {
		t.Fatalf("body = %q", got)
	}
}

func TestRequestOptionsBodyValueWins(t *testing.T) {
	form := js.FormData().New()
	form.Call("append", "k", "v")

	init := RequestOptions{Method: "POST", Body: "ignored", BodyValue: form}.apply()

	body := init.Get("body")
	if body.Type() == js.TypeString {
		t.Fatalf("BodyValue lost to Body: %q", body.String())
	}
	if got := body.Call("get", "k").String(); got != "v" {
		t.Fatalf("form field = %q", got)
	}
}

func TestRequestOptionsEmptyBodyIsUnset(t *testing.T) {
	init := RequestOptions{}.apply()
	if init.Get("body").Type() != js.TypeUndefined {
		t.Fatal("empty options set a body")
	}
	if init.Get("method").Type() != js.TypeUndefined {
		t.Fatal("empty options set a method")
	}
}
