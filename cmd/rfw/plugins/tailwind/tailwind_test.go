package tailwind

import "testing"

// TestShouldRebuild ensures the plugin's rebuild triggers are detected
// correctly based on file paths and extensions.
func TestShouldRebuild(t *testing.T) {
	p := &plugin{output: "tailwind.css"}

	if !p.ShouldRebuild("style.css") {
		t.Fatalf("expected css change to trigger rebuild")
	}
	if p.ShouldRebuild("tailwind.css") {
		t.Fatalf("output file should not trigger rebuild")
	}
	if !p.ShouldRebuild("index.html") || !p.ShouldRebuild("tmpl.rtml") {
		t.Fatalf("html/rtml should trigger rebuild")
	}
	if !p.ShouldRebuild("main.go") {
		t.Fatalf("go files should trigger rebuild")
	}
	if p.ShouldRebuild("image.png") {
		t.Fatalf("unrelated files should not trigger rebuild")
	}
}
