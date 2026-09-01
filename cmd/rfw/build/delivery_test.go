//go:build !js

package build

import (
	"strings"
	"testing"
)

// A manifest written before build.delivery existed keeps building exactly as it
// did, and a value that is not a delivery mode fails the build with something
// the author can act on.
func TestParseDelivery(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  delivery
	}{
		{name: "absent", value: "", want: deliveryNetwork},
		{name: "explicit network", value: "network", want: deliveryNetwork},
		{name: "embedded", value: "embedded", want: deliveryEmbedded},
		{name: "surrounding whitespace", value: "  embedded\n", want: deliveryEmbedded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDelivery(tc.value)
			if err != nil {
				t.Fatalf("parseDelivery(%q): %v", tc.value, err)
			}
			if got != tc.want {
				t.Fatalf("parseDelivery(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestParseDeliveryRejectsUnknownModes(t *testing.T) {
	for _, value := range []string{"Embedded", "local", "offline", "true", "net work"} {
		_, err := parseDelivery(value)
		if err == nil {
			t.Fatalf("parseDelivery(%q) was accepted", value)
		}
		message := err.Error()
		for _, want := range []string{"build.delivery", value, string(deliveryNetwork), string(deliveryEmbedded)} {
			if !strings.Contains(message, want) {
				t.Fatalf("error %q does not mention %q", message, want)
			}
		}
	}
}

func TestParseSSCTransport(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  sscTransport
	}{
		{value: "", want: sscTransportBrowser},
		{value: "browser", want: sscTransportBrowser},
		{value: " capacitor ", want: sscTransportCapacitor},
	} {
		got, err := parseSSCTransport(tc.value)
		if err != nil {
			t.Fatalf("parseSSCTransport(%q): %v", tc.value, err)
		}
		if got != tc.want {
			t.Fatalf("parseSSCTransport(%q) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestParseSSCTransportRejectsUnknownModes(t *testing.T) {
	for _, value := range []string{"native", "Capacitor", "auto", "true"} {
		_, err := parseSSCTransport(value)
		if err == nil {
			t.Fatalf("parseSSCTransport(%q) was accepted", value)
		}
		message := err.Error()
		for _, want := range []string{"build.sscTransport", value, string(sscTransportBrowser), string(sscTransportCapacitor)} {
			if !strings.Contains(message, want) {
				t.Fatalf("error %q does not mention %q", message, want)
			}
		}
	}
}

// Delivery mode and type are separate decisions. An embedded SSC application
// still links the host client and still configures its WebSocket endpoint; an
// embedded static application still drops both.
func TestDecodeBuildShapeKeepsDeliveryAndTypeSeparate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		manifest string
		want     buildShape
	}{
		{
			name:     "ssc defaults to network delivery",
			manifest: `{"build":{"type":"ssc","host":"wss://api.example.com/ws"}}`,
			want:     buildShape{host: "wss://api.example.com/ws", delivery: deliveryNetwork, transport: sscTransportBrowser},
		},
		{
			name:     "embedded ssc keeps the host client and the host url",
			manifest: `{"build":{"type":"ssc","host":"wss://api.example.com/ws","delivery":"embedded"}}`,
			want:     buildShape{host: "wss://api.example.com/ws", delivery: deliveryEmbedded, transport: sscTransportBrowser},
		},
		{
			name:     "embedded static stays static",
			manifest: `{"build":{"type":"static","delivery":"embedded"}}`,
			want:     buildShape{static: true, delivery: deliveryEmbedded, transport: sscTransportBrowser},
		},
		{
			name:     "an empty manifest is a network build",
			manifest: `{}`,
			want:     buildShape{delivery: deliveryNetwork, transport: sscTransportBrowser},
		},
		{
			name:     "embedded SSC can select the Capacitor transport",
			manifest: `{"build":{"type":"ssc","host":"wss://api.example.com/ws","delivery":"embedded","sscTransport":"capacitor"}}`,
			want:     buildShape{host: "wss://api.example.com/ws", delivery: deliveryEmbedded, transport: sscTransportCapacitor},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeBuildShape([]byte(tc.manifest))
			if err != nil {
				t.Fatalf("decodeBuildShape: %v", err)
			}
			if got.static != tc.want.static {
				t.Fatalf("static = %t, want %t", got.static, tc.want.static)
			}
			if got.host != tc.want.host {
				t.Fatalf("host = %q, want %q", got.host, tc.want.host)
			}
			if got.delivery != tc.want.delivery {
				t.Fatalf("delivery = %q, want %q", got.delivery, tc.want.delivery)
			}
			if got.transport != tc.want.transport {
				t.Fatalf("transport = %q, want %q", got.transport, tc.want.transport)
			}
		})
	}
}

func TestDecodeBuildShapeRejectsAnUnknownDelivery(t *testing.T) {
	if _, err := decodeBuildShape([]byte(`{"build":{"type":"ssc","delivery":"capacitor"}}`)); err == nil {
		t.Fatal("an unknown delivery mode built without an error")
	}
}

func TestDecodeBuildShapeRejectsAnUnknownSSCTransport(t *testing.T) {
	if _, err := decodeBuildShape([]byte(`{"build":{"type":"ssc","sscTransport":"native"}}`)); err == nil {
		t.Fatal("an unknown SSC transport built without an error")
	}
}

// An unreadable manifest has always fallen back to the defaults rather than
// failing the build, and it still does.
func TestDecodeBuildShapeToleratesAnUnparsableManifest(t *testing.T) {
	got, err := decodeBuildShape([]byte("not json"))
	if err != nil {
		t.Fatalf("decodeBuildShape: %v", err)
	}
	if got.delivery != deliveryNetwork || got.transport != sscTransportBrowser || got.static || got.plugins != nil {
		t.Fatalf("shape = %+v, want the defaults", got)
	}
}

// Negotiation needs a live host on the other end. A static origin has none, and
// packaged assets come from a local asset handler that negotiates nothing.
func TestBuildShapeNegotiates(t *testing.T) {
	for _, tc := range []struct {
		shape buildShape
		want  bool
	}{
		{shape: buildShape{delivery: deliveryNetwork}, want: true},
		{shape: buildShape{static: true, delivery: deliveryNetwork}, want: false},
		{shape: buildShape{delivery: deliveryEmbedded}, want: false},
		{shape: buildShape{static: true, delivery: deliveryEmbedded}, want: false},
	} {
		if got := tc.shape.negotiates(); got != tc.want {
			t.Fatalf("%+v negotiates = %t, want %t", tc.shape, got, tc.want)
		}
	}
}
