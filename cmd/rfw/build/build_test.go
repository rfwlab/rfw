//go:build !js

package build

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

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

func TestCompressWasmBrotli(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	src := "app.wasm"
	content := []byte(strings.Repeat("rfw wasm", 32))
	if err := os.WriteFile(src, content, 0o600); err != nil {
		t.Fatalf("write wasm: %v", err)
	}

	if err := compressWasmBrotli(src); err != nil {
		t.Fatalf("compressWasmBrotli: %v", err)
	}

	brPath := src + ".br"
	f, err := os.Open(brPath)
	if err != nil {
		t.Fatalf("open brotli file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("close brotli file: %v", err)
		}
	}()

	reader := brotli.NewReader(f)
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read brotli: %v", err)
	}
	if string(decompressed) != string(content) {
		t.Fatalf("unexpected decompressed content")
	}
}

func TestWriteClientConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := writeClientConfig(".", "wss://example.com/rfw", "abc123"); err != nil {
		t.Fatalf("writeClientConfig: %v", err)
	}

	config, err := os.ReadFile("rfw_config.js")
	if err != nil {
		t.Fatalf("read client config: %v", err)
	}
	got := string(config)
	if !strings.Contains(got, `window.RFW_HOST_URL = "wss://example.com/rfw";`) {
		t.Fatalf("host URL missing from client config: %q", got)
	}
	if !strings.Contains(got, `window.RFW_WASM_VERSION = "abc123";`) {
		t.Fatalf("wasm version missing from client config: %q", got)
	}
}
