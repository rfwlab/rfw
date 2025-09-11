package utils

import (
	"bytes"
	"io"
	"net/http/httptest"
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

func TestIsDebug(t *testing.T) {
	EnableDebug(true)
	if !IsDebug() {
		t.Fatalf("expected true in debug mode")
	}
	EnableDebug(false)
	if IsDebug() {
		t.Fatalf("expected false when debug disabled")
	}
}

func TestPrintStartupInfo(t *testing.T) {
	out := captureOutput(func() { PrintStartupInfo("8080", "8443", "192.168.0.1", true) })
	if !strings.Contains(out, "http://localhost:8080/") {
		t.Fatalf("expected local URL in output, got %q", out)
	}
	if !strings.Contains(out, "http://192.168.0.1:8080/") {
		t.Fatalf("expected network URL, got %q", out)
	}
	out = captureOutput(func() { PrintStartupInfo("8080", "8443", "", false) })
	if !strings.Contains(out, "--host") {
		t.Fatalf("expected hint about --host, got %q", out)
	}
}

func TestPrintHelp(t *testing.T) {
	out := captureOutput(PrintHelp)
	if !strings.Contains(out, "Shortcuts") || !strings.Contains(out, "Flags") {
		t.Fatalf("missing help sections, got %q", out)
	}
}

func TestLogServeRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo", nil)
	out := captureOutput(func() { LogServeRequest(req) })
	if !strings.Contains(out, "/foo") {
		t.Fatalf("expected path in output, got %q", out)
	}
}
