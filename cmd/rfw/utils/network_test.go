package utils

import (
	"os"
	"testing"
)

func TestOpenBrowserError(t *testing.T) {
	orig := os.Getenv("BROWSER")
	_ = os.Setenv("BROWSER", "nonexistent-browser")
	defer os.Setenv("BROWSER", orig)

	if err := OpenBrowser("http://example.com"); err == nil {
		t.Fatalf("expected error when browser command is missing")
	}
}
