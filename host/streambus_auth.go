package host

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"sync"
	"time"
)

const (
	streamBusAuthPath    = "/__rfw/streambus-auth"
	streamBusAuthTTL     = 30 * time.Second
	streamBusAuthLimit   = 256
	streamBusAuthHeaders = 16 << 10
	streamBusAuthTimeout = 5 * time.Second
)

type streamBusCredential struct {
	request *http.Request
	origin  string
	expires time.Time
}

type streamBusAuth struct {
	mu      sync.Mutex
	tickets map[string]streamBusCredential
}

func (e *streamBusEndpoint) issueCredential(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	typeName, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if typeName != "application/json" || r.ContentLength != 0 || r.URL.RawQuery != "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	origin := r.Header.Get("Origin")
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	allowed := origin != "" && origin == scheme+"://"+r.Host
	for _, configured := range e.runtime.origins {
		allowed = allowed || origin != "" && origin == configured
	}
	if !allowed {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}
	if e.runtime.authFunc == nil {
		http.Error(w, "authentication is not configured", http.StatusServiceUnavailable)
		return
	}
	headerBytes := 0
	for key, values := range r.Header {
		headerBytes += len(key)
		for _, value := range values {
			headerBytes += len(value)
		}
	}
	if headerBytes > streamBusAuthHeaders {
		http.Error(w, "request headers too large", http.StatusRequestHeaderFieldsTooLarge)
		return
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		http.Error(w, "authentication unavailable", http.StatusServiceUnavailable)
		return
	}
	token := hex.EncodeToString(secret[:])
	request := r.Clone(context.WithoutCancel(r.Context()))
	request.Body = http.NoBody
	request.GetBody = nil
	request.ContentLength = 0
	now := time.Now()
	e.auth.mu.Lock()
	for key, ticket := range e.auth.tickets {
		if !now.Before(ticket.expires) {
			delete(e.auth.tickets, key)
		}
	}
	if len(e.auth.tickets) >= streamBusAuthLimit {
		e.auth.mu.Unlock()
		http.Error(w, "authentication capacity reached", http.StatusTooManyRequests)
		return
	}
	e.auth.tickets[token] = streamBusCredential{request: request, origin: origin, expires: now.Add(streamBusAuthTTL)}
	e.auth.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (e *streamBusEndpoint) consumeCredential(request *http.Request, token string) (*http.Request, error) {
	if len(token) != 64 {
		return nil, errors.New("streambus: authentication refused")
	}
	e.auth.mu.Lock()
	credential, ok := e.auth.tickets[token]
	delete(e.auth.tickets, token)
	e.auth.mu.Unlock()
	if !ok || !time.Now().Before(credential.expires) || credential.origin != request.Header.Get("Origin") {
		return nil, errors.New("streambus: authentication refused")
	}
	// Recheck the original credential at consumption; issuance is not a grant
	// that survives server-side logout or revocation.
	authenticated := request.Clone(context.WithoutCancel(credential.request.Context()))
	authenticated.Header = credential.request.Header.Clone()
	authenticated.Body = http.NoBody
	authenticated.GetBody = nil
	if e.runtime.authFunc == nil || !e.runtime.authFunc(authenticated) {
		return nil, errors.New("streambus: authentication refused")
	}
	return authenticated, nil
}

func decodeStreamBusCredential(payload []byte) (string, error) {
	var frame struct {
		Token string `json:"token"`
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&frame); err != nil {
		return "", errors.New("streambus: authentication refused")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return "", errors.New("streambus: authentication refused")
	}
	return frame.Token, nil
}
