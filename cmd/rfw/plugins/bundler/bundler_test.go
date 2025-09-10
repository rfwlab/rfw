package bundler

import (
	"os"
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
	if !p.ShouldRebuild("index.html") {
		t.Fatalf("expected html change to trigger rebuild")
	}
	if p.ShouldRebuild(filepath.Join("build", "app.js")) {
		t.Fatalf("output directory should not trigger rebuild")
	}
	if p.ShouldRebuild("image.png") {
		t.Fatalf("unrelated files should not trigger rebuild")
	}
}

func TestIsTailwindCSS(t *testing.T) {
	dir := t.TempDir()
	tw := filepath.Join(dir, "input.css")
	if err := os.WriteFile(tw, []byte("@import \"tailwindcss\";"), 0o644); err != nil {
		t.Fatalf("write tailwind file: %v", err)
	}
	if !isTailwindCSS(tw) {
		t.Fatalf("expected tailwind directives to be detected")
	}

	normal := filepath.Join(dir, "normal.css")
	if err := os.WriteFile(normal, []byte("body{}"), 0o644); err != nil {
		t.Fatalf("write normal file: %v", err)
	}
	if isTailwindCSS(normal) {
		t.Fatalf("unexpected tailwind detection in normal css")
	}
}

func TestPostBuildMinifiesFiles(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build", "static")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		t.Fatalf("mkdir build: %v", err)
	}

	jsFile := filepath.Join(buildDir, "app.js")
	if err := os.WriteFile(jsFile, []byte("function add ( a , b ){ return a + b ; }"), 0o644); err != nil {
		t.Fatalf("write js: %v", err)
	}
	cssFile := filepath.Join(buildDir, "app.css")
	if err := os.WriteFile(cssFile, []byte("body { color: red; }"), 0o644); err != nil {
		t.Fatalf("write css: %v", err)
	}
	htmlFile := filepath.Join(buildDir, "index.html")
	html := "<html><head><title> hi </title></head><body> <h1> hi </h1> </body></html>"
	if err := os.WriteFile(htmlFile, []byte(html), 0o644); err != nil {
		t.Fatalf("write html: %v", err)
	}

	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	p := &plugin{}
	if err := p.PostBuild(nil); err != nil {
		t.Fatalf("postbuild: %v", err)
	}

	outJS, _ := os.ReadFile(jsFile)
	if len(outJS) >= len("function add ( a , b ){ return a + b ; }") {
		t.Fatalf("js not minified: %s", outJS)
	}
	outCSS, _ := os.ReadFile(cssFile)
	if len(outCSS) >= len("body { color: red; }") {
		t.Fatalf("css not minified: %s", outCSS)
	}
	outHTML, _ := os.ReadFile(htmlFile)
	if len(outHTML) >= len(html) {
		t.Fatalf("html not minified: %s", outHTML)
	}
}
