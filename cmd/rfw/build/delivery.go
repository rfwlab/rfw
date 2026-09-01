//go:build !js

// Package build compiles RFW applications and runs build plugins. This file
// covers how the built client reaches the browser: over the network from a host
// or a static origin, or from assets packaged next to the application.
package build

import (
	"encoding/json"
	"fmt"
	"strings"
)

// delivery names how the client bundle reaches the browser. It is orthogonal to
// build.type: type decides what is linked into the bundle, delivery decides how
// the bundle is transferred.
type delivery string

// sscTransport names the runtime used for the client-to-host SSC connection.
// It is independent of delivery: an embedded application may still use the
// browser WebSocket, while a Capacitor application can opt into the native
// URLSession bridge when its authenticated cookie cannot cross WKWebView's
// origin boundary.
type sscTransport string

const (
	// deliveryNetwork is the default. The bundle is downloaded, so the build
	// writes compressed artifacts and the loader picks one.
	deliveryNetwork delivery = "network"
	// deliveryEmbedded packages the bundle with the application, as a native
	// container does. The packaged file is still read through the container's
	// local URL-scheme asset handler, but there is no network transfer to
	// compress and nothing that negotiates a content coding, so the build
	// writes only the optimized raw wasm and the loader loads exactly that.
	deliveryEmbedded delivery = "embedded"

	sscTransportBrowser   sscTransport = "browser"
	sscTransportCapacitor sscTransport = "capacitor"
)

// parseDelivery resolves the build.delivery value from rfw.json. An empty value
// keeps the network default, so a manifest written before this setting existed
// builds exactly as it did.
func parseDelivery(value string) (delivery, error) {
	switch mode := delivery(strings.TrimSpace(value)); mode {
	case "":
		return deliveryNetwork, nil
	case deliveryNetwork, deliveryEmbedded:
		return mode, nil
	default:
		return "", fmt.Errorf(
			"rfw.json: build.delivery %q is not a delivery mode; use %q "+
				"(the default: compressed artifacts fetched over the network) or %q "+
				"(the raw bundle packaged with the application)",
			value, deliveryNetwork, deliveryEmbedded,
		)
	}
}

func parseSSCTransport(value string) (sscTransport, error) {
	switch transport := sscTransport(strings.TrimSpace(value)); transport {
	case "", sscTransportBrowser:
		return sscTransportBrowser, nil
	case sscTransportCapacitor:
		return transport, nil
	default:
		return "", fmt.Errorf(
			"rfw.json: build.sscTransport %q is not an SSC transport; use %q "+
				"(the default browser WebSocket) or %q (the RFW Capacitor plugin)",
			value, sscTransportBrowser, sscTransportCapacitor,
		)
	}
}

// buildShape is the delivery-relevant view of rfw.json. It exists so the two
// settings that shape a build stay separable: "static" decides whether the SSC
// client is linked, "embedded" decides how the bundle is delivered, and neither
// implies the other.
type buildShape struct {
	static    bool
	host      string
	delivery  delivery
	transport sscTransport
	plugins   map[string]json.RawMessage
}

// negotiates reports whether the client can leave the choice of content coding
// to the server. A static origin does not negotiate, and the local asset
// handler that serves packaged assets does not either.
func (s buildShape) negotiates() bool {
	return !s.static && s.delivery == deliveryNetwork
}

// compresses reports whether the build should write compressed artifacts.
// Packaging them would ship bytes nothing will ever request.
func (s buildShape) compresses() bool {
	return s.delivery == deliveryNetwork
}

// decodeBuildShape reads rfw.json. A manifest that does not parse falls back to
// the defaults, as it always has, but a manifest that parses and names an
// unknown delivery mode fails the build rather than quietly ignoring it.
func decodeBuildShape(data []byte) (buildShape, error) {
	var manifest struct {
		Build struct {
			Type         string `json:"type"`
			Host         string `json:"host"`
			Delivery     string `json:"delivery"`
			SSCTransport string `json:"sscTransport"`
		} `json:"build"`
		Plugins map[string]json.RawMessage `json:"plugins"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return buildShape{delivery: deliveryNetwork, transport: sscTransportBrowser}, nil
	}
	mode, err := parseDelivery(manifest.Build.Delivery)
	if err != nil {
		return buildShape{}, err
	}
	transport, err := parseSSCTransport(manifest.Build.SSCTransport)
	if err != nil {
		return buildShape{}, err
	}
	return buildShape{
		static:    manifest.Build.Type == "static",
		host:      manifest.Build.Host,
		delivery:  mode,
		transport: transport,
		plugins:   manifest.Plugins,
	}, nil
}
