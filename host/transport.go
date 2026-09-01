package host

import (
	"encoding/json"
	"os"
	"strings"
)

// Transport identifies the browser-to-host transport used by SSC.
type Transport string

const (
	TransportWebSocket Transport = "websocket"
	TransportStreamBus Transport = "streambus"
	TransportAuto      Transport = "auto"
)

// ResolveTransport reads RFW_TRANSPORT first, then the top-level transport
// key in rfw.json. Unknown and empty values preserve the WebSocket default.
func ResolveTransport() Transport {
	if mode := normalizeTransport(os.Getenv("RFW_TRANSPORT")); mode != "" {
		return mode
	}
	var manifest struct {
		Transport string `json:"transport"`
	}
	if data, err := os.ReadFile("rfw.json"); err == nil {
		if json.Unmarshal(data, &manifest) == nil {
			if mode := normalizeTransport(manifest.Transport); mode != "" {
				return mode
			}
		}
	}
	return TransportWebSocket
}

func normalizeTransport(value string) Transport {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(TransportWebSocket), "ws":
		return TransportWebSocket
	case string(TransportStreamBus), "webtransport", "warp-streambus":
		return TransportStreamBus
	case string(TransportAuto):
		return TransportAuto
	default:
		return ""
	}
}

func streamBusEnabled() bool {
	mode := ResolveTransport()
	return mode == TransportStreamBus || mode == TransportAuto
}
