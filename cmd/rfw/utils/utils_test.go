package utils

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput redirects stdout for the duration of f and returns what was
// written to it.
func captureOutput(f func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestDebug(t *testing.T) {
	EnableDebug(true)
	out := captureOutput(func() { Debug("hello") })
	if !strings.Contains(out, "[rfw][debug]") {
		t.Fatalf("expected debug output, got %q", out)
	}

	EnableDebug(false)
	out = captureOutput(func() { Debug("no output") })
	if out != "" {
		t.Fatalf("expected no output, got %q", out)
	}
}
