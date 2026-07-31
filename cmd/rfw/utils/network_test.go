//go:build !js

package utils

import "testing"

func TestOpenBrowserError(t *testing.T) {
	t.Setenv("BROWSER", "nonexistent-browser")

	if err := OpenBrowser("http://example.com"); err == nil {
		t.Fatalf("expected error when browser command is missing")
	}
}
