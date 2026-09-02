//go:build js && wasm

package hostclient

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	stdjs "syscall/js"
)

const (
	hostTransportWebSocket = "websocket"
	hostTransportStreamBus = "streambus"
	hostTransportAuto      = "auto"
	streamBusPath          = "/streambus"
)

func preferredHostTransport() string {
	value := strings.ToLower(strings.TrimSpace(stdjs.Global().Get("RFW_TRANSPORT").String()))
	switch value {
	case hostTransportStreamBus, "webtransport", "warp-streambus":
		return hostTransportStreamBus
	case hostTransportAuto:
		return hostTransportAuto
	default:
		return hostTransportWebSocket
	}
}

func hostStreamBusURL() string {
	if configured := stdjs.Global().Get("RFW_STREAMBUS_URL"); configured.Truthy() {
		return normalizeStreamBusURL(configured.String(), false)
	}
	return normalizeStreamBusURL(hostWSURL(), true)
}

func normalizeStreamBusURL(raw string, incrementHTTPPort bool) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(raw, "wss://"):
		raw = "https://" + strings.TrimPrefix(raw, "wss://")
		incrementHTTPPort = false
	case strings.HasPrefix(raw, "ws://"):
		raw = "https://" + strings.TrimPrefix(raw, "ws://")
	case strings.HasPrefix(raw, "http://"):
		raw = "https://" + strings.TrimPrefix(raw, "http://")
	case strings.HasPrefix(raw, "https://"):
		incrementHTTPPort = false
	default:
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	if incrementHTTPPort {
		if port, err := strconv.Atoi(parsed.Port()); err == nil && port > 0 {
			parsed.Host = net.JoinHostPort(parsed.Hostname(), strconv.Itoa(port+1))
		}
	}
	if parsed.Path == "" || parsed.Path == "/" || parsed.Path == "/ws" {
		parsed.Path = streamBusPath
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func dialStreamBus(ctx context.Context, endpoint string) (*hostConn, error) {
	constructor := stdjs.Global().Get("WebTransport")
	if constructor.Type() != stdjs.TypeFunction {
		return nil, errors.New("hostclient: WebTransport API is unavailable")
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
		transport.Call("close")
		return nil, err
	}
	stream, err := awaitPromise(ctx, transport.Call("createBidirectionalStream"))
	if err != nil {
		transport.Call("close")
		return nil, err
	}
	reader := &jsReader{reader: stream.Get("readable").Call("getReader")}
	writer := &jsWriter{writer: stream.Get("writable").Call("getWriter")}
	buffered := bufio.NewReader(reader)
	c := newHostConn()
	c.send = func(data string) error {
		return writeStreamBusFrame(context.Background(), writer, []byte(data))
	}
	c.shutdown = func() error {
		reader.reader.Call("cancel")
		writer.writer.Call("close")
		transport.Call("close")
		return nil
	}
	c.openOnce.Do(func() { close(c.open) })
	go func() {
		for {
			payload, readErr := readStreamBusFrame(buffered, int(maxInboundMessageBytes))
			if readErr != nil {
				c.fail(readErr)
				return
			}
			c.deliver(payload)
		}
	}()
	return c, nil
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
		return streamBusClientConfig{options: stdjs.Undefined()}, errors.New("hostclient: invalid StreamBus certificate hash")
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

func writeStreamBusFrame(ctx context.Context, writer *jsWriter, payload []byte) error {
	var prefix [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(prefix[:], uint64(len(payload)))
	if err := writer.write(ctx, prefix[:n]); err != nil {
		return err
	}
	return writer.write(ctx, payload)
}

func readStreamBusFrame(reader *bufio.Reader, maximum int) ([]byte, error) {
	size, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, err
	}
	if maximum > 0 && size > uint64(maximum) {
		return nil, fmt.Errorf("hostclient: StreamBus frame of %d bytes exceeds %d", size, maximum)
	}
	if size > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("hostclient: StreamBus frame of %d bytes exceeds platform limit", size)
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
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
