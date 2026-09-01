//go:build js && wasm

package hostclient

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	stdjs "syscall/js"

	rfwjs "github.com/rfwlab/rfw/v2/js"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type serverWireMessage struct {
	Component   string       `json:"component"`
	Action      string       `json:"action"`
	Control     string       `json:"control"`
	ID          string       `json:"id"`
	Payload     any          `json:"payload"`
	Error       *ActionError `json:"error"`
	Session     string       `json:"session"`
	Sequence    uint64       `json:"sequence"`
	Ack         uint64       `json:"ack"`
	ResumeToken string       `json:"resumeToken"`
}

type transportConnection interface {
	Name() string
	Write(context.Context, wireMessage) error
	Read(context.Context, *serverWireMessage) error
	Ping(context.Context) error
	Close()
}

type webSocketTransport struct{ connection *websocket.Conn }

func (c *webSocketTransport) Name() string { return "websocket" }
func (c *webSocketTransport) Write(ctx context.Context, message wireMessage) error {
	return wsjson.Write(ctx, c.connection, message)
}
func (c *webSocketTransport) Read(ctx context.Context, message *serverWireMessage) error {
	return wsjson.Read(ctx, c.connection, message)
}
func (c *webSocketTransport) Ping(ctx context.Context) error { return c.connection.Ping(ctx) }
func (c *webSocketTransport) Close() {
	_ = c.connection.Close(websocket.StatusNormalClosure, "")
}

func preferredTransport() string {
	value := strings.ToLower(strings.TrimSpace(rfwjs.Get("RFW_TRANSPORT").String()))
	switch value {
	case "streambus", "webtransport", "warp-streambus":
		return "streambus"
	case "auto":
		return "auto"
	default:
		return "websocket"
	}
}

func dialPreferredTransport(ctx context.Context) (transportConnection, error) {
	mode := preferredTransport()
	if mode == "streambus" || mode == "auto" {
		connection, err := dialStreamBus(ctx, hostStreamBusURL())
		if err == nil {
			return connection, nil
		}
		if debug {
			log.Printf("hostclient: StreamBus unavailable, falling back to WebSocket: %v", err)
		}
	}
	connection, _, err := websocket.Dial(ctx, hostWSURL(), nil)
	if err != nil {
		return nil, err
	}
	return &webSocketTransport{connection: connection}, nil
}

func hostStreamBusURL() string {
	if configured := rfwjs.Get("RFW_STREAMBUS_URL"); configured.Truthy() {
		return normalizeStreamBusURL(configured.String(), false)
	}
	if configured := rfwjs.Get("RFW_HOST_URL"); configured.Truthy() {
		return normalizeStreamBusURL(configured.String(), true)
	}
	host := browserLocationHost()
	if configured := rfwjs.Get("RFW_HOST"); configured.Truthy() {
		host = configured.String()
	}
	return normalizeStreamBusURL(host, browserLocationProtocol() == "http:")
}

func browserLocationHost() string {
	window := stdjs.Global().Get("window")
	if window.Type() == stdjs.TypeUndefined || window.Type() == stdjs.TypeNull {
		return "localhost"
	}
	location := window.Get("location")
	if location.Type() == stdjs.TypeUndefined || location.Type() == stdjs.TypeNull {
		return "localhost"
	}
	return location.Get("host").String()
}

func browserLocationProtocol() string {
	window := stdjs.Global().Get("window")
	if window.Type() == stdjs.TypeUndefined || window.Type() == stdjs.TypeNull {
		return "http:"
	}
	location := window.Get("location")
	if location.Type() == stdjs.TypeUndefined || location.Type() == stdjs.TypeNull {
		return "http:"
	}
	return location.Get("protocol").String()
}

func normalizeStreamBusURL(raw string, incrementHTTPPort bool) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(raw, "wss://"):
		raw = "https://" + strings.TrimPrefix(raw, "wss://")
	case strings.HasPrefix(raw, "ws://"):
		raw = "https://" + strings.TrimPrefix(raw, "ws://")
		incrementHTTPPort = true
	case strings.HasPrefix(raw, "http://"):
		raw = "https://" + strings.TrimPrefix(raw, "http://")
		incrementHTTPPort = true
	case strings.HasPrefix(raw, "https://"):
	default:
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if incrementHTTPPort {
		hostname := parsed.Hostname()
		if port, err := strconv.Atoi(parsed.Port()); err == nil && port > 0 {
			parsed.Host = fmt.Sprintf("%s:%d", hostname, port+1)
		}
	}
	if parsed.Path == "" || parsed.Path == "/" || parsed.Path == "/ws" {
		parsed.Path = "/streambus"
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

type streamBusTransport struct {
	transport stdjs.Value
	reader    *jsReader
	writer    *jsWriter
	buffered  *bufio.Reader
}

func dialStreamBus(ctx context.Context, endpoint string) (*streamBusTransport, error) {
	constructor := stdjs.Global().Get("WebTransport")
	if constructor.Type() != stdjs.TypeFunction {
		return nil, errors.New("WebTransport API is unavailable")
	}
	config, _ := loadStreamBusConfig(ctx)
	if config.port != "" {
		endpoint = replaceURLPort(endpoint, config.port)
	}
	var transport stdjs.Value
	if config.options.Type() == stdjs.TypeObject {
		transport = constructor.New(endpoint, config.options)
	} else {
		transport = constructor.New(endpoint)
	}
	if _, err := awaitPromise(ctx, transport.Get("ready")); err != nil {
		return nil, err
	}
	bidirectional, err := awaitPromise(ctx, transport.Call("createBidirectionalStream"))
	if err != nil {
		transport.Call("close")
		return nil, err
	}
	reader := &jsReader{reader: bidirectional.Get("readable").Call("getReader")}
	writer := &jsWriter{writer: bidirectional.Get("writable").Call("getWriter")}
	return &streamBusTransport{transport: transport, reader: reader, writer: writer, buffered: bufio.NewReader(reader)}, nil
}

type streamBusClientConfig struct {
	options stdjs.Value
	port    string
}

func loadStreamBusConfig(ctx context.Context) (streamBusClientConfig, error) {
	fetch := stdjs.Global().Get("fetch")
	if fetch.Type() != stdjs.TypeFunction {
		return streamBusClientConfig{options: stdjs.Undefined()}, nil
	}
	response, err := awaitPromise(ctx, fetch.Invoke("/__rfw/streambus-config"))
	if err != nil || !response.Get("ok").Bool() {
		return streamBusClientConfig{options: stdjs.Undefined()}, err
	}
	config, err := awaitPromise(ctx, response.Call("json"))
	if err != nil {
		return streamBusClientConfig{options: stdjs.Undefined()}, err
	}
	hash, err := hex.DecodeString(config.Get("certificateHash").String())
	if err != nil || len(hash) != 32 {
		return streamBusClientConfig{options: stdjs.Undefined()}, err
	}
	value := stdjs.Global().Get("Uint8Array").New(len(hash))
	stdjs.CopyBytesToJS(value, hash)
	descriptor := stdjs.Global().Get("Object").New()
	descriptor.Set("algorithm", "sha-256")
	descriptor.Set("value", value.Get("buffer"))
	list := stdjs.Global().Get("Array").New()
	list.Call("push", descriptor)
	options := stdjs.Global().Get("Object").New()
	options.Set("serverCertificateHashes", list)
	return streamBusClientConfig{options: options, port: config.Get("port").String()}, nil
}

func replaceURLPort(raw, port string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" || port == "" {
		return raw
	}
	parsed.Host = net.JoinHostPort(parsed.Hostname(), port)
	return parsed.String()
}

func (c *streamBusTransport) Name() string { return "warp-streambus" }
func (c *streamBusTransport) Write(ctx context.Context, message wireMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	var prefix [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(prefix[:], uint64(len(payload)))
	if err := c.writer.write(ctx, prefix[:n]); err != nil {
		return err
	}
	return c.writer.write(ctx, payload)
}
func (c *streamBusTransport) Read(ctx context.Context, message *serverWireMessage) error {
	size, err := binary.ReadUvarint(c.buffered)
	if err != nil {
		return err
	}
	if size > 64<<20 {
		return fmt.Errorf("hostclient: StreamBus frame too large: %d", size)
	}
	payload := make([]byte, int(size))
	if _, err := io.ReadFull(c.buffered, payload); err != nil {
		return err
	}
	return json.Unmarshal(payload, message)
}
func (c *streamBusTransport) Ping(context.Context) error { return nil }
func (c *streamBusTransport) Close() {
	c.reader.reader.Call("cancel")
	c.writer.writer.Call("close")
	c.transport.Call("close")
}

type jsReader struct {
	reader stdjs.Value
	buffer []byte
}

func (r *jsReader) Read(payload []byte) (int, error) {
	if len(r.buffer) == 0 {
		result, err := awaitPromise(context.Background(), r.reader.Call("read"))
		if err != nil {
			return 0, err
		}
		if result.Get("done").Bool() {
			return 0, io.EOF
		}
		value := result.Get("value")
		r.buffer = make([]byte, value.Get("byteLength").Int())
		stdjs.CopyBytesToGo(r.buffer, value)
	}
	n := copy(payload, r.buffer)
	r.buffer = r.buffer[n:]
	return n, nil
}

type jsWriter struct{ writer stdjs.Value }

func (w *jsWriter) write(ctx context.Context, payload []byte) error {
	array := stdjs.Global().Get("Uint8Array").New(len(payload))
	stdjs.CopyBytesToJS(array, payload)
	_, err := awaitPromise(ctx, w.writer.Call("write", array))
	return err
}

type promiseResult struct {
	value stdjs.Value
	err   error
}

func awaitPromise(ctx context.Context, promise stdjs.Value) (stdjs.Value, error) {
	result := make(chan promiseResult, 1)
	resolve := stdjs.FuncOf(func(_ stdjs.Value, args []stdjs.Value) any {
		value := stdjs.Undefined()
		if len(args) > 0 {
			value = args[0]
		}
		result <- promiseResult{value: value}
		return nil
	})
	reject := stdjs.FuncOf(func(_ stdjs.Value, args []stdjs.Value) any {
		message := "promise rejected"
		if len(args) > 0 {
			message = args[0].String()
		}
		result <- promiseResult{err: errors.New(message)}
		return nil
	})
	promise.Call("then", resolve).Call("catch", reject)
	select {
	case settled := <-result:
		resolve.Release()
		reject.Release()
		return settled.value, settled.err
	case <-ctx.Done():
		go func() {
			<-result
			resolve.Release()
			reject.Release()
		}()
		return stdjs.Undefined(), ctx.Err()
	}
}
