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

type Request struct {
ID     string          `json:"id"`
Method string          `json:"method"`
Params json.RawMessage `json:"params,omitempty"`
}

type Response struct {
ID     string          `json:"id"`
Result json.RawMessage `json:"result,omitempty"`
Error  string          `json:"error,omitempty"`
}

type HandlerFunc func(context.Context, json.RawMessage) (any, error)

type Server struct {
mu       sync.RWMutex
handlers map[string]HandlerFunc
}

func NewServer() *Server { return &Server{handlers: map[string]HandlerFunc{}} }

func (s *Server) Register(method string, h HandlerFunc) {
s.mu.Lock()
defer s.mu.Unlock()
s.handlers[method] = h
}

var ErrMethodNotFound = errors.New("rpc: method not found")

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
defer resp.Body.Close()

var res Response
if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
return err
}
if res.Error != "" {
return fmt.Errorf("rpc: %s", res.Error)
}
if out == nil {
return nil
}
return json.Unmarshal(res.Result, out)
}
