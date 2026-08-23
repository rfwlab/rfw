//go:build !js

package build

import (
	"compress/gzip"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

func TestBrowserLoaderScenarios(t *testing.T) {
	_, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required to exercise the browser loader")
	}
	loader, err := filepath.Abs(filepath.Join("..", "initproj", "template", "wasm_loader.js"))
	if err != nil {
		t.Fatalf("resolve loader: %v", err)
	}
	script, err := filepath.Abs(filepath.Join("testdata", "wasm_loader_test.mjs"))
	if err != nil {
		t.Fatalf("resolve loader test: %v", err)
	}
	// #nosec G204 -- both script paths are resolved from this repository.
	cmd := exec.Command("node", script, loader)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("browser loader scenarios failed: %v\n%s", err, output)
	}
}

// TestCopyFile ensures copyFile replicates the source file's contents at the
// destination path.
func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	src := "src.txt"
	dst := "dst.txt"
	content := []byte("hello world")
	if err := os.WriteFile(src, content, 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("expected %q, got %q", content, got)
	}
}

func TestRemoveStaleArtifacts(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	src := "app.wasm"
	artifacts := []string{src + ".br", src + ".gz"}
	for _, path := range artifacts {
		if err := os.WriteFile(path, []byte("stale"), 0o600); err != nil {
			t.Fatalf("write stale %s: %v", path, err)
		}
	}

	if err := removeStaleArtifacts(src); err != nil {
		t.Fatalf("removeStaleArtifacts: %v", err)
	}
	for _, path := range artifacts {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected %s removed, stat err=%v", path, err)
		}
	}

	// Absence of the files must be a no-op, not an error.
	if err := removeStaleArtifacts(src); err != nil {
		t.Fatalf("removeStaleArtifacts on missing files: %v", err)
	}
}

// A production build writes every artifact it will later advertise, and each
// one has to round-trip: a truncated or mislabelled artifact would be served
// under an immutable cache header.
func TestCompressWasmWritesEveryArtifact(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	src := "app.wasm"
	content := []byte(strings.Repeat("rfw wasm", 32))
	if err := os.WriteFile(src, content, 0o600); err != nil {
		t.Fatalf("write wasm: %v", err)
	}

	encodings, err := compressWasm(src)
	if err != nil {
		t.Fatalf("compressWasm: %v", err)
	}
	if len(encodings) != 2 || encodings[0] != EncodingBrotli || encodings[1] != EncodingGzip {
		t.Fatalf("encodings = %v, want brotli then gzip", encodings)
	}

	decoders := map[Encoding]func(io.Reader) (io.Reader, error){
		EncodingBrotli: func(r io.Reader) (io.Reader, error) { return brotli.NewReader(r), nil },
		EncodingGzip:   func(r io.Reader) (io.Reader, error) { return gzip.NewReader(r) },
	}
	for _, encoding := range encodings {
		path := src + artifactExt[encoding]
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open %s: %v", path, err)
		}
		decoder, err := decoders[encoding](f)
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		decompressed, err := io.ReadAll(decoder)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if err := f.Close(); err != nil {
			t.Errorf("close %s: %v", path, err)
		}
		if string(decompressed) != string(content) {
			t.Fatalf("%s did not round-trip", path)
		}
	}
}

func TestWriteClientConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	encodings := []Encoding{EncodingBrotli, EncodingGzip}
	if err := writeClientConfig(".", "wss://example.com/rfw", "abc123", encodings, true, true); err != nil {
		t.Fatalf("writeClientConfig: %v", err)
	}

	config, err := os.ReadFile("rfw_config.js")
	if err != nil {
		t.Fatalf("read client config: %v", err)
	}
	got := string(config)
	for _, want := range []string{
		`window.RFW_HOST_URL = "wss://example.com/rfw";`,
		`window.RFW_WASM_VERSION = "abc123";`,
		`window.RFW_WASM_ENCODINGS = ["br", "gzip"];`,
		`window.RFW_WASM_NEGOTIATED = true;`,
		`window.RFW_BUILD_MODE = "production";`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("client config is missing %q:\n%s", want, got)
		}
	}
}

// A build that produced no compressed artifact must say so. An empty list is a
// fact the loader can act on; an absent global is indistinguishable from an
// older build and is what let an application silently serve the raw bundle.
func TestWriteClientConfigDescribesAnUncompressedBuild(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := writeClientConfig(".", "", "devbuild", nil, false, false); err != nil {
		t.Fatalf("writeClientConfig: %v", err)
	}
	config, err := os.ReadFile("rfw_config.js")
	if err != nil {
		t.Fatalf("read client config: %v", err)
	}
	got := string(config)
	for _, want := range []string{
		`window.RFW_WASM_ENCODINGS = [];`,
		`window.RFW_WASM_NEGOTIATED = false;`,
		`window.RFW_BUILD_MODE = "development";`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("client config is missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "RFW_HOST_URL") {
		t.Fatalf("an unset host emitted a host URL:\n%s", got)
	}
}

// Every global the loader reads has to be defined by the config, or the loader
// cannot tell a disabled feature from a stale build.
func TestWriteClientConfigDefinesEveryGlobalTheLoaderReads(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := writeClientConfig(".", "", "abc123", []Encoding{EncodingGzip}, false, true); err != nil {
		t.Fatalf("writeClientConfig: %v", err)
	}
	config, err := os.ReadFile("rfw_config.js")
	if err != nil {
		t.Fatalf("read client config: %v", err)
	}
	loader, err := os.ReadFile(filepath.Join("..", "..", "initproj", "template", "wasm_loader.js"))
	if err != nil {
		t.Skipf("loader template unavailable: %v", err)
	}
	defined := string(config)
	for _, global := range regexp.MustCompile(`window\.(RFW_[A-Z_]+)`).FindAllStringSubmatch(string(loader), -1) {
		if !strings.Contains(defined, "window."+global[1]+" =") {
			t.Fatalf("the loader reads window.%s but the config never defines it", global[1])
		}
	}
}

func TestStampBootstrapScripts(t *testing.T) {
	cases := map[string]struct{ html, want string }{
		"unversioned": {
			html: `<script src="/rfw_config.js"></script>`,
			want: `<script src="/rfw_config.js?v=abc123"></script>`,
		},
		"stale version is replaced": {
			html: `<script src="/rfw_config.js?v=2"></script>`,
			want: `<script src="/rfw_config.js?v=abc123"></script>`,
		},
		"relative path": {
			html: `<script src="wasm_exec.js"></script>`,
			want: `<script src="wasm_exec.js?v=abc123"></script>`,
		},
		"extra attributes are preserved": {
			html: `<script defer src="/wasm_loader.js" data-x="1"></script>`,
			want: `<script defer src="/wasm_loader.js?v=abc123" data-x="1"></script>`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, changed := stampBootstrapScripts(tc.html, "abc123")
			if !changed {
				t.Fatalf("stamping reported no change for %q", tc.html)
			}
			if got != tc.want {
				t.Fatalf("stamped = %q, want %q", got, tc.want)
			}
		})
	}
}

// Application markup the build does not recognise is left exactly as written.
// Rewriting it would be worse than not versioning it.
func TestStampBootstrapScriptsLeavesUnrelatedMarkupAlone(t *testing.T) {
	for _, html := range []string{
		`<script src="/vendor/analytics.js"></script>`,
		`<script>const go = new Go();</script>`,
		`<link rel="stylesheet" href="/rfw_config.js" />`,
	} {
		got, changed := stampBootstrapScripts(html, "abc123")
		if changed || got != html {
			t.Fatalf("stamping rewrote unrelated markup %q into %q", html, got)
		}
	}
}

// A release has to be able to invalidate a cached loader, so every bootstrap
// script in the built page carries the build version.
func TestStampIndexHTMLVersionsEveryBootstrapScript(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	page := `<html><body><div id="app"></div>` +
		`<script src="/rfw_config.js?v=2"></script>` +
		`<script src="/wasm_exec.js"></script>` +
		`<script src="/wasm_loader.js"></script>` +
		`</body></html>`
	if err := os.WriteFile("index.html", []byte(page), 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := stampIndexHTML("index.html", "deadbeef"); err != nil {
		t.Fatalf("stampIndexHTML: %v", err)
	}
	stamped, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	for _, script := range versionedScripts {
		want := script + "?v=deadbeef"
		if !strings.Contains(string(stamped), want) {
			t.Fatalf("index.html is missing %q:\n%s", want, stamped)
		}
	}
	if strings.Contains(string(stamped), "?v=2") {
		t.Fatalf("a stale version survived the stamp:\n%s", stamped)
	}
}

// A missing page is not an error: a static build may ship no index.html.
func TestStampIndexHTMLIgnoresAMissingPage(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := stampIndexHTML("index.html", "abc123"); err != nil {
		t.Fatalf("stampIndexHTML on a missing page: %v", err)
	}
}

// TestWriteSSCImport checks the generated file carries the wasm build tag and
// the hostclient blank import, and that it parses as Go. core no longer
// imports hostclient, so this file is what keeps SSC registration wired in a
// non-static build.
func TestWriteSSCImport(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	path, err := writeSSCImport()
	if err != nil {
		t.Fatalf("writeSSCImport error: %v", err)
	}
	if path != sscImportFile {
		t.Fatalf("expected %q, got %q", sscImportFile, path)
	}

	data, err := os.ReadFile("rfw_ssc_gen.go")
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	src := string(data)
	if !strings.HasPrefix(src, "//go:build js && wasm\n") {
		t.Fatalf("generated file must open with the wasm build tag, got:\n%s", src)
	}
	if !strings.Contains(src, `import _ "github.com/rfwlab/rfw/v2/hostclient"`) {
		t.Fatalf("generated file is missing the hostclient blank import:\n%s", src)
	}
	if !strings.Contains(src, "DO NOT EDIT.") {
		t.Fatalf("generated file is missing the generated marker:\n%s", src)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), path, data, parser.SkipObjectResolution); err != nil {
		t.Fatalf("generated file does not parse: %v", err)
	}
}
