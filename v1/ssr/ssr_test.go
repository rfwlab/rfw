package ssr

import "testing"

func TestRender(t *testing.T) {
	tpl := []byte("<p>Hello {{name}}</p>")
	out, err := Render(tpl, map[string]any{"name": "Alice"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	expected := "<p>Hello Alice</p><script id=\"__RFWDATA__\" type=\"application/json\">{\"name\":\"Alice\"}</script>"
	if out != expected {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestRenderWithSpaces(t *testing.T) {
	tpl := []byte("<p>Hello {{ name }}</p>")
	out, err := Render(tpl, map[string]any{"name": "Bob"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	expected := "<p>Hello Bob</p><script id=\"__RFWDATA__\" type=\"application/json\">{\"name\":\"Bob\"}</script>"
	if out != expected {
		t.Fatalf("unexpected output: %s", out)
	}
}
