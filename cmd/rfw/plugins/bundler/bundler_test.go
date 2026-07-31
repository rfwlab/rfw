//go:build !js

package bundler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rfwlab/rfw/v2/cmd/rfw/utils"
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
	if !isTailwindCSSData([]byte("@import \"tailwindcss\";")) {
		t.Fatalf("expected tailwind directives to be detected")
	}
	if isTailwindCSSData([]byte("body{}")) {
		t.Fatalf("unexpected tailwind detection in normal css")
	}
}

func TestPostBuildMinifiesFiles(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build", "static")
	if err := os.MkdirAll(buildDir, 0o750); err != nil {
		t.Fatalf("mkdir build: %v", err)
	}

	jsFile := filepath.Join(buildDir, "app.js")
	if err := os.WriteFile(jsFile, []byte("function add ( a , b ){ return a + b ; }"), 0o600); err != nil {
		t.Fatalf("write js: %v", err)
	}
	cssFile := filepath.Join(buildDir, "app.css")
	if err := os.WriteFile(cssFile, []byte("body { color: red; }"), 0o600); err != nil {
		t.Fatalf("write css: %v", err)
	}
	htmlFile := filepath.Join(buildDir, "index.html")
	html := "<html><head><title> hi </title></head><body> <h1> hi </h1> </body></html>"
	if err := os.WriteFile(htmlFile, []byte(html), 0o600); err != nil {
		t.Fatalf("write html: %v", err)
	}

	t.Chdir(dir)

	utils.EnableDebug(false)

	p := &plugin{}
	if err := p.PostBuild(nil); err != nil {
		t.Fatalf("postbuild: %v", err)
	}

	outJS, err := os.ReadFile("build/static/app.js")
	if err != nil {
		t.Fatal(err)
	}
	if len(outJS) >= len("function add ( a , b ){ return a + b ; }") {
		t.Fatalf("js not minified: %s", outJS)
	}
	outCSS, err := os.ReadFile("build/static/app.css")
	if err != nil {
		t.Fatal(err)
	}
	if len(outCSS) >= len("body { color: red; }") {
		t.Fatalf("css not minified: %s", outCSS)
	}
	outHTML, err := os.ReadFile("build/static/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if len(outHTML) >= len(html) {
		t.Fatalf("html not minified: %s", outHTML)
	}
}

func TestPostBuildSkippedInDebug(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build", "static")
	if err := os.MkdirAll(buildDir, 0o750); err != nil {
		t.Fatalf("mkdir build: %v", err)
	}

	jsFile := filepath.Join(buildDir, "app.js")
	content := "function add ( a , b ){ return a + b ; }"
	if err := os.WriteFile(jsFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write js: %v", err)
	}

	t.Chdir(dir)

	utils.EnableDebug(true)
	defer utils.EnableDebug(false)

	p := &plugin{}
	if err := p.PostBuild(nil); err != nil {
		t.Fatalf("postbuild: %v", err)
	}

	out, err := os.ReadFile("build/static/app.js")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != content {
		t.Fatalf("file should remain unminified, got %s", out)
	}
}
