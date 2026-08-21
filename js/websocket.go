//go:build js && wasm

package js

import (
	"errors"
	"fmt"
)

// Browser WebSocket readyState values.
const (
	SocketConnecting = 0
	SocketOpen       = 1
	SocketClosing    = 2
	SocketClosed     = 3
)

// ErrSocketNotOpen reports a send attempted on a socket that is not open.
var ErrSocketNotOpen = errors.New("js: websocket is not open")

var socketEvents = [...]string{"onopen", "onmessage", "onerror", "onclose"}

// SocketHandlers receives the events of a browser WebSocket. Every callback
// runs on the JavaScript event loop, so none of them may block.
type SocketHandlers struct {
	Open  func()
	Text  func(string)
	Bytes func([]byte)
	Error func(error)
	Close func(code int, reason string, clean bool)
	// MaxMessageBytes reports an inbound payload above the limit through Error
	// instead of delivering it. Zero accepts any size.
	MaxMessageBytes int
}

// Socket is a browser WebSocket together with the callbacks bound to it.
type Socket struct {
	value     Value
	callbacks []Func
	released  bool
}

// OpenSocket connects to url and binds handlers to the browser WebSocket. The
// returned socket is still connecting: wait for Open before sending.
func OpenSocket(url string, handlers SocketHandlers) (*Socket, error) {
	var value Value
	if err := socketCall(func() { value = WebSocket().New(url) }); err != nil {
		return nil, err
	}
	s := &Socket{value: value}
	// The host sends binary frames, which arrive as Blob unless the socket is
	// told to hand them over as ArrayBuffer.
	s.value.Set("binaryType", "arraybuffer")
	s.bind("onopen", func(Value) {
		if handlers.Open != nil {
			handlers.Open()
		}
	})
	s.bind("onmessage", func(event Value) { s.deliver(event, handlers) })
	s.bind("onerror", func(Value) {
		if handlers.Error != nil {
			handlers.Error(errors.New("js: websocket error"))
		}
	})
	s.bind("onclose", func(event Value) {
		if handlers.Close == nil {
			return
		}
		handlers.Close(event.Get("code").Int(), event.Get("reason").String(), event.Get("wasClean").Bool())
	})
	return s, nil
}

// SendText writes data as a text frame.
func (s *Socket) SendText(data string) error {
	if s.ReadyState() != SocketOpen {
		return ErrSocketNotOpen
	}
	return socketCall(func() { s.value.Call("send", data) })
}

// Close closes the socket with the given code and reason and releases its
// callbacks. The browser only accepts 1000 or the 3000-4999 application range.
func (s *Socket) Close(code int, reason string) error {
	var err error
	if state := s.ReadyState(); state == SocketConnecting || state == SocketOpen {
		err = socketCall(func() { s.value.Call("close", code, reason) })
	}
	s.Release()
	return err
}

// ReadyState reports the current browser readyState.
func (s *Socket) ReadyState() int {
	if s == nil || !s.value.Truthy() {
		return SocketClosed
	}
	return s.value.Get("readyState").Int()
}

// Release detaches every callback and frees the Go functions behind them. A
// released socket delivers no further events.
func (s *Socket) Release() {
	if s.released {
		return
	}
	s.released = true
	for _, event := range socketEvents {
		s.value.Set(event, Null())
	}
	for _, callback := range s.callbacks {
		callback.Release()
	}
	s.callbacks = nil
}

func (s *Socket) bind(event string, fn func(Value)) {
	callback := SafeFuncOf(func(_ Value, args []Value) any {
		var value Value
		if len(args) > 0 {
			value = args[0]
		}
		fn(value)
		return nil
	})
	s.callbacks = append(s.callbacks, callback)
	s.value.Set(event, callback)
}

func (s *Socket) deliver(event Value, handlers SocketHandlers) {
	data := event.Get("data")
	if data.Type() == TypeString {
		// A JavaScript string reports UTF-16 units, so bound the conversion
		// first and only then measure the encoded payload.
		if handlers.oversize(data.Get("length").Int()) {
			return
		}
		text := data.String()
		if handlers.oversize(len(text)) {
			return
		}
		if handlers.Text != nil {
			handlers.Text(text)
		}
		return
	}
	size := data.Get("byteLength").Int()
	if handlers.oversize(size) {
		return
	}
	payload := make([]byte, size)
	CopyBytesToGo(payload, Uint8Array().New(data))
	if handlers.Bytes != nil {
		handlers.Bytes(payload)
	}
}

func (h SocketHandlers) oversize(size int) bool {
	if h.MaxMessageBytes <= 0 || size <= h.MaxMessageBytes {
		return false
	}
	if h.Error != nil {
		h.Error(fmt.Errorf("js: websocket message of %d bytes exceeds the %d byte limit", size, h.MaxMessageBytes))
	}
	return true
}

func socketCall(fn func()) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("js: websocket call failed: %v", recovered)
		}
	}()
	fn()
	return nil
}
