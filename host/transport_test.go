package host

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveTransportPrecedence(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skip("os.Chdir is not implemented on js")
	}
	dir := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })

	if got := ResolveTransport(); got != TransportWebSocket {
		t.Fatalf("default = %q", got)
	}
	if err := os.WriteFile(filepath.Join(dir, "rfw.json"), []byte(`{"transport":"streambus"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := ResolveTransport(); got != TransportStreamBus {
		t.Fatalf("manifest = %q", got)
	}
	t.Setenv("RFW_TRANSPORT", "ws")
	if got := ResolveTransport(); got != TransportWebSocket {
		t.Fatalf("environment = %q", got)
	}
}

func TestResolveTransportAliases(t *testing.T) {
	for value, want := range map[string]Transport{
		"websocket":      TransportWebSocket,
		"WS":             TransportWebSocket,
		"streambus":      TransportStreamBus,
		"webtransport":   TransportStreamBus,
		"warp-streambus": TransportStreamBus,
		"auto":           TransportAuto,
	} {
		if got := normalizeTransport(value); got != want {
			t.Errorf("normalizeTransport(%q) = %q, want %q", value, got, want)
		}
	}
}
