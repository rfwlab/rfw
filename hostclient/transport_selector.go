//go:build js && wasm

package hostclient

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

const (
	sscTransportBrowser   = "browser"
	sscTransportCapacitor = "capacitor"
	capacitorPluginName   = "RFWSSC"
)

var nativeConnectionSequence atomic.Uint64

// dial selects the transport stamped into rfw_config.js. Browser WebSocket is
// the compatibility default. A configured native transport never falls back:
// silently doing so would restore the cookie boundary the native transport was
// selected to avoid.
func dial(ctx context.Context, url string) (*hostConn, error) {
	switch sscTransport() {
	case sscTransportBrowser:
		return dialBrowser(ctx, url)
	case sscTransportCapacitor:
		return dialCapacitor(ctx, url)
	default:
		return nil, fmt.Errorf("hostclient: unsupported SSC transport %q", sscTransport())
	}
}

func sscTransport() string {
	configured := js.Get("RFW_SSC_TRANSPORT")
	if !configured.Truthy() {
		return sscTransportBrowser
	}
	return strings.TrimSpace(configured.String())
}

// dialCapacitor connects through the RFW Capacitor plugin. Authentication
// cookies and the Origin header are owned by the native URLSession; Go/WASM
// receives protocol frames and lifecycle events, never cookie values.
func dialCapacitor(ctx context.Context, url string) (*hostConn, error) {
	plugin, err := capacitorSSCPlugin()
	if err != nil {
		return nil, err
	}

	c := newHostConn()
	id := fmt.Sprintf("rfw-ssc-%d", nativeConnectionSequence.Add(1))
	var callback js.Func
	var releaseOnce sync.Once
	releaseCallback := func() { releaseOnce.Do(callback.Release) }
	releaseAfterCallback := func() { time.AfterFunc(time.Millisecond, releaseCallback) }
	callback = js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 1 && args[1].Truthy() {
			c.fail(errors.New("hostclient: native SSC callback failed"))
			releaseAfterCallback()
			return nil
		}
		if len(args) == 0 || !args[0].Truthy() {
			c.fail(errors.New("hostclient: native SSC callback returned no event"))
			releaseAfterCallback()
			return nil
		}
		event := args[0]
		switch event.Get("type").String() {
		case "open":
			c.openOnce.Do(func() { close(c.open) })
		case "message":
			payload, decodeErr := decodeNativeFrame(event)
			if decodeErr != nil {
				c.fail(decodeErr)
				return nil
			}
			c.deliver(payload)
		case "error":
			c.fail(nativeEventError(event, "native SSC transport failed"))
			releaseAfterCallback()
		case "close":
			c.fail(fmt.Errorf(
				"hostclient: native SSC connection closed with code %d: %s",
				event.Get("code").Int(), event.Get("reason").String(),
			))
			releaseAfterCallback()
		default:
			c.fail(errors.New("hostclient: native SSC transport returned an unknown event"))
			releaseAfterCallback()
		}
		return nil
	})
	// Native close crosses the Capacitor bridge asynchronously. Keep the Go
	// callback alive long enough for the final close/error event; a terminal
	// callback releases it sooner through the same once guard.
	c.release = func() { time.AfterFunc(time.Second, releaseCallback) }
	c.send = func(data string) error {
		options := js.NewDict()
		options.Set("id", id)
		options.Set("data", data)
		return callNativePlugin(plugin, "send", options.Value)
	}
	c.shutdown = func() error {
		options := js.NewDict()
		options.Set("id", id)
		return callNativePlugin(plugin, "close", options.Value)
	}

	options := js.NewDict()
	options.Set("id", id)
	options.Set("url", url)
	if err := callNativePlugin(plugin, "connect", options.Value, callback); err != nil {
		_ = c.close()
		return nil, err
	}

	select {
	case <-c.open:
		return c, nil
	case <-c.done:
		_ = c.close()
		return nil, c.err()
	case <-ctx.Done():
		_ = c.close()
		return nil, ctx.Err()
	}
}

func capacitorSSCPlugin() (js.Value, error) {
	capacitor := js.Get("Capacitor")
	if !capacitor.Truthy() {
		return js.Undefined(), errors.New("hostclient: Capacitor is unavailable for native SSC transport")
	}
	plugins := capacitor.Get("Plugins")
	if !plugins.Truthy() {
		return js.Undefined(), errors.New("hostclient: Capacitor plugins are unavailable for native SSC transport")
	}
	plugin := plugins.Get(capacitorPluginName)
	if !plugin.Truthy() {
		return js.Undefined(), errors.New("hostclient: RFWSSC Capacitor plugin is not installed")
	}
	return plugin, nil
}

func decodeNativeFrame(event js.Value) ([]byte, error) {
	data := event.Get("data").String()
	switch event.Get("encoding").String() {
	case "text":
		return []byte(data), nil
	case "base64":
		payload, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, errors.New("hostclient: native SSC returned invalid base64")
		}
		return payload, nil
	default:
		return nil, errors.New("hostclient: native SSC returned an unknown frame encoding")
	}
}

func nativeEventError(event js.Value, fallback string) error {
	message := strings.TrimSpace(event.Get("message").String())
	if message == "" {
		message = fallback
	}
	return errors.New("hostclient: " + message)
}

func callNativePlugin(plugin js.Value, method string, args ...any) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("hostclient: native SSC %s call failed", method)
		}
	}()
	plugin.Call(method, args...)
	return nil
}
