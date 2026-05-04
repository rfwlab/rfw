package rpc

import (
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"testing"
)

func TestServerHandle_DispatchSuccess(t *testing.T) {
s := NewServer()
s.Register("echo", func(_ context.Context, p json.RawMessage) (any, error) {
var in struct {
S string `json:"s"`
}
_ = json.Unmarshal(p, &in)
return map[string]string{"out": in.S}, nil
})

req := []byte(`{"id":"99","method":"echo","params":{"s":"hi"}}`)
b, err := s.Handle(context.Background(), req)
if err != nil {
t.Fatalf("Handle error: %v", err)
}

var res Response
if err := json.Unmarshal(b, &res); err != nil {
t.Fatalf("unmarshal response: %v", err)
}
if res.ID != "99" || res.Error != "" {
t.Fatalf("unexpected response: %+v", res)
}
var out map[string]string
if err := json.Unmarshal(res.Result, &out); err != nil {
t.Fatalf("unmarshal result: %v", err)
}
if out["out"] != "hi" {
t.Fatalf("expected hi, got %v", out)
}
}

func TestServerHandle_MethodNotFound(t *testing.T) {
s := NewServer()
b, err := s.Handle(context.Background(), []byte(`{"id":"1","method":"missing"}`))
if err != nil {
t.Fatalf("Handle error: %v", err)
}
var res Response
_ = json.Unmarshal(b, &res)
if res.Error == "" {
t.Fatalf("expected error")
}
}

func TestCall_HTTPRoundTrip(t *testing.T) {
s := NewServer()
s.Register("add", func(_ context.Context, p json.RawMessage) (any, error) {
var in struct {
A int `json:"a"`
B int `json:"b"`
}
_ = json.Unmarshal(p, &in)
return map[string]int{"sum": in.A + in.B}, nil
})

srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
var req Request
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
t.Fatalf("decode request: %v", err)
}
b, err := s.Handle(r.Context(), mustJSON(req))
if err != nil {
t.Fatalf("handle: %v", err)
}
w.Header().Set("Content-Type", "application/json")
_, _ = w.Write(b)
}))
defer srv.Close()

var out struct {
Sum int `json:"sum"`
}
if err := Call(context.Background(), srv.URL, "add", map[string]int{"a": 2, "b": 5}, &out); err != nil {
t.Fatalf("Call error: %v", err)
}
if out.Sum != 7 {
t.Fatalf("expected 7, got %d", out.Sum)
}
}

func mustJSON(v any) []byte {
b, _ := json.Marshal(v)
return b
}
