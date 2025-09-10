package bundler

import (
	"path/filepath"
	"testing"
)

func TestShouldRebuild(t *testing.T) {
	p := &plugin{}
	if !p.ShouldRebuild("main.js") {
		t.Fatalf("expected js change to trigger rebuild")
	}
	if !p.ShouldRebuild("styles.css") {
		t.Fatalf("expected css change to trigger rebuild")
	}
	if !p.ShouldRebuild("page.rtml") {
		t.Fatalf("expected rtml change to trigger rebuild")
	}
	if p.ShouldRebuild(filepath.Join("build", "app.js")) {
		t.Fatalf("output directory should not trigger rebuild")
	}
	if p.ShouldRebuild("image.png") {
		t.Fatalf("unrelated files should not trigger rebuild")
	}
}
