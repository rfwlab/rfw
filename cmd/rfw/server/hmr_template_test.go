package server

import (
	"path/filepath"
	"testing"
)

func TestComponentNamesForTemplate(t *testing.T) {
	tmpl := filepath.Join("..", "..", "..", "docs", "examples", "components", "templates", "input_component.rtml")
	names := componentNamesForTemplate(tmpl)
	t.Logf("names: %v", names)
	if len(names) == 0 {
		t.Fatalf("expected component names for %s", tmpl)
	}
	found := false
	for _, name := range names {
		if name == "InputComponent" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected InputComponent in %v", names)
	}
}

func TestComponentNamesForTemplateGenerics(t *testing.T) {
	tmpl := filepath.Join("..", "..", "..", "docs", "examples", "components", "templates", "webgl_component.rtml")
	names := componentNamesForTemplate(tmpl)
	t.Logf("names: %v", names)
	found := false
	for _, name := range names {
		if name == "WebGLComponent" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected WebGLComponent in %v", names)
	}
}

func TestComponentNamesForTemplateNoMatch(t *testing.T) {
	tmpl := filepath.Join("hmr_template_test.go")
	if names := componentNamesForTemplate(tmpl); names != nil {
		t.Fatalf("expected nil, got %v", names)
	}
}
