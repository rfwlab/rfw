package build

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

// TestCopyFile ensures copyFile replicates the source file's contents at the
// destination path.
func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	content := []byte("hello world")
	if err := os.WriteFile(src, content, 0o644); err != nil {
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
	src := filepath.Join(dir, "app.wasm")
	content := []byte(strings.Repeat("rfw wasm", 32))
	if err := os.WriteFile(src, content, 0o644); err != nil {
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
	defer f.Close()

	reader := brotli.NewReader(f)
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read brotli: %v", err)
	}
	if string(decompressed) != string(content) {
		t.Fatalf("unexpected decompressed content")
	}
}
