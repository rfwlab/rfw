//go:build !js || !wasm

package safehttp

import (
	"context"
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"
)

func TestIsPublic(t *testing.T) {
	tests := map[string]bool{
		"8.8.8.8":              true,
		"2606:4700:4700::1111": true,
		"127.0.0.1":            false,
		"10.0.0.1":             false,
		"100.64.0.1":           false,
		"169.254.169.254":      false,
		"192.168.1.1":          false,
		"::1":                  false,
		"fd00::1":              false,
		"fe80::1":              false,
	}
	for rawAddress, want := range tests {
		if got := isPublic(netip.MustParseAddr(rawAddress)); got != want {
			t.Errorf("isPublic(%q) = %t, want %t", rawAddress, got, want)
		}
	}
}

func TestNewRequestRejectsUnsafeURLs(t *testing.T) {
	for _, rawURL := range []string{
		"/relative",
		"file:///tmp/secret",
		"http://user:pass@example.com",
	} {
		if _, err := NewRequest(context.Background(), "GET", rawURL); err == nil {
			t.Errorf("expected %q to be rejected", rawURL)
		}
	}
}

func TestPublicDialRejectsLoopback(t *testing.T) {
	dial := publicDialContext(
		&net.Dialer{Timeout: time.Second},
		net.DefaultResolver,
	)
	if conn, err := dial(context.Background(), "tcp", "127.0.0.1:80"); err == nil {
		_ = conn.Close()
		t.Fatal("expected loopback address to be rejected")
	} else if !strings.Contains(err.Error(), "not public") {
		t.Fatalf("unexpected rejection error: %v", err)
	}
}
