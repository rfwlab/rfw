//go:build !js

package utils

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{name: "minor with two digits", current: "v2.9.0", latest: "v2.10.0", want: true},
		{name: "stable after prerelease", current: "v2.1.0-beta.19", latest: "v2.1.0", want: true},
		{name: "prerelease before stable", current: "v2.1.0", latest: "v2.1.0-beta.19"},
		{name: "numeric prerelease", current: "v2.1.0-beta.9", latest: "v2.1.0-beta.10", want: true},
		{name: "build metadata ignored", current: "v2.1.0+one", latest: "v2.1.0+two"},
		{name: "older release", current: "v3.0.0", latest: "v2.99.99"},
		{name: "invalid current", current: "development", latest: "v2.1.0"},
		{name: "invalid latest", current: "v2.1.0", latest: "v2.01.0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isNewer(test.current, test.latest); got != test.want {
				t.Fatalf("isNewer(%q, %q) = %v, want %v", test.current, test.latest, got, test.want)
			}
		})
	}
}

func TestCopyWithSHA256(t *testing.T) {
	content := []byte("verified update")
	expected := sha256.Sum256(content)
	var destination bytes.Buffer

	if err := copyWithSHA256(&destination, bytes.NewReader(content), expected); err != nil {
		t.Fatalf("copy verified content: %v", err)
	}
	if !bytes.Equal(destination.Bytes(), content) {
		t.Fatalf("copied %q, want %q", destination.Bytes(), content)
	}

	destination.Reset()
	wrong := sha256.Sum256([]byte("different"))
	if err := copyWithSHA256(&destination, bytes.NewReader(content), wrong); err == nil {
		t.Fatal("expected checksum mismatch")
	}
}

func TestParseSHA256Digest(t *testing.T) {
	content := sha256.Sum256([]byte("release"))
	digest := "sha256:" + fmt.Sprintf("%x", content)
	got, err := parseSHA256Digest(digest)
	if err != nil {
		t.Fatalf("parse digest: %v", err)
	}
	if got != content {
		t.Fatalf("parsed digest %x, want %x", got, content)
	}
	if _, err := parseSHA256Digest(""); err == nil {
		t.Fatal("expected missing digest error")
	}
}

// captureOutput redirects stdout for the duration of f and returns what was
// written to it.
func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	f()
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	os.Stdout = orig
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("close reader: %v", err)
		}
	}()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestDebug(t *testing.T) {
	EnableDebug(true)
	out := captureOutput(t, func() { Debug("hello") })
	if !strings.Contains(out, "[rfw][debug]") {
		t.Fatalf("expected debug output, got %q", out)
	}

	EnableDebug(false)
	out = captureOutput(t, func() { Debug("no output") })
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
	out := captureOutput(t, func() { PrintStartupInfo("8080", "8443", "192.168.0.1", true) })
	if !strings.Contains(out, "http://localhost:8080/") {
		t.Fatalf("expected local URL in output, got %q", out)
	}
	if !strings.Contains(out, "http://192.168.0.1:8080/") {
		t.Fatalf("expected network URL, got %q", out)
	}
	out = captureOutput(t, func() { PrintStartupInfo("8080", "8443", "", false) })
	if !strings.Contains(out, "--host") {
		t.Fatalf("expected hint about --host, got %q", out)
	}
}

func TestPrintHelp(t *testing.T) {
	out := captureOutput(t, PrintHelp)
	if !strings.Contains(out, "Shortcuts") || !strings.Contains(out, "Flags") {
		t.Fatalf("missing help sections, got %q", out)
	}
}

func TestLogServeRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo", nil)
	out := captureOutput(t, func() { LogServeRequest(req) })
	if !strings.Contains(out, "/foo") {
		t.Fatalf("expected path in output, got %q", out)
	}
}

func TestReplaceExecutable(t *testing.T) {
	dir := t.TempDir()
	source := dir + "/source"
	target := dir + "/target"
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Errorf("close test root: %v", err)
		}
	})
	if err := root.WriteFile("source", []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := root.WriteFile("target", []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := replaceExecutable(source, target); err != nil {
		t.Fatalf("replace executable: %v", err)
	}
	content, err := root.ReadFile("target")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Fatalf("target contains %q, want %q", content, "new")
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
}
