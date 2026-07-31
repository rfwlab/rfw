// Package rpc provides JSON request handling and HTTP RPC calls.
package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// Request is a JSON RPC request.
type Request struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON RPC response.
type Response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// HandlerFunc handles an RPC method.
type HandlerFunc func(context.Context, json.RawMessage) (any, error)

// Server dispatches requests to registered handlers.
type Server struct {
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

// NewServer creates an empty RPC server.
func NewServer() *Server { return &Server{handlers: map[string]HandlerFunc{}} }

// Register adds or replaces a method handler.
func (s *Server) Register(method string, h HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = h
}

// ErrMethodNotFound is returned for an unknown RPC method.
var ErrMethodNotFound = errors.New("rpc: method not found")

// Handle decodes and dispatches one RPC request.
func (s *Server) Handle(ctx context.Context, reqBytes []byte) ([]byte, error) {
	var req Request
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		return nil, err
	}
	if req.Method == "" {
		return nil, errors.New("rpc: missing method")
	}

	s.mu.RLock()
	h := s.handlers[req.Method]
	s.mu.RUnlock()

	if h == nil {
		res := Response{ID: req.ID, Error: ErrMethodNotFound.Error()}
		return json.Marshal(res)
	}

	out, err := h(ctx, req.Params)
	if err != nil {
		res := Response{ID: req.ID, Error: err.Error()}
		return json.Marshal(res)
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	res := Response{ID: req.ID, Result: b}
	return json.Marshal(res)
}

// Call invokes an RPC method over HTTP.
func Call(ctx context.Context, endpoint, method string, params any, out any) error {
	req := Request{ID: "1", Method: method}
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return err
		}
		req.Params = b
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}

	var res Response
	decodeErr := json.NewDecoder(resp.Body).Decode(&res)
	closeErr := resp.Body.Close()
	if decodeErr != nil {
		return decodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if res.Error != "" {
		return fmt.Errorf("rpc: %s", res.Error)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(res.Result, out)
}
