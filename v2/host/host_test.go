package host

import (
	"testing"
)

// TestHostComponent verifies registration and handler execution.
func TestHostComponent(t *testing.T) {
	called := false
	hc := NewHostComponent("cmp", func(payload map[string]any) any {
		called = true
		if payload["x"] != 1 {
			t.Fatalf("unexpected payload: %v", payload)
		}
		return "ok"
	})
	Register(hc)
	got, ok := Get("cmp")
	if !ok || got != hc {
		t.Fatalf("component not registered")
	}
	if resp := hc.Handle(map[string]any{"x": 1}); resp != "ok" || !called {
		t.Fatalf("handler not executed or wrong response: %v", resp)
	}
}

func TestHostComponentWithSession(t *testing.T) {
	hc := NewHostComponentWithSession("withSession", func(session *Session, payload map[string]any) any {
		if session == nil {
			t.Fatalf("session should not be nil")
		}
		store := session.StoreManager().NewStore("test")
		store.Set("value", payload["v"])
		return store.Snapshot()
	})

	sess := newSession("abc")
	resp := hc.HandleWithSession(sess, map[string]any{"v": 42})
	snap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("unexpected response type %T", resp)
	}
	if snap["value"] != 42 {
		t.Fatalf("unexpected store snapshot: %v", snap)
	}
	if !hc.SessionAware() {
		t.Fatalf("expected session aware component")
	}
	if hc.StoreManager(sess) != sess.StoreManager() {
		t.Fatalf("StoreManager helper mismatch")
	}
}

// TestLogLevel checks environment variable parsing.
func TestLogLevel(t *testing.T) {
	t.Setenv("RFW_LOG_LEVEL", "debug")
	if lvl := logLevel(); lvl.String() != "DEBUG" {
		t.Fatalf("expected DEBUG level, got %s", lvl)
	}
	t.Setenv("RFW_LOG_LEVEL", "warn")
	if lvl := logLevel(); lvl.String() != "WARN" {
		t.Fatalf("expected WARN level, got %s", lvl)
	}
	t.Setenv("RFW_LOG_LEVEL", "")
	if lvl := logLevel(); lvl.String() != "INFO" {
		t.Fatalf("expected INFO level, got %s", lvl)
	}
}

// TestGenerateSelfSignedCert ensures a certificate is generated.
func TestGenerateSelfSignedCert(t *testing.T) {
	cert, err := generateSelfSignedCert()
	if err != nil {
		t.Fatalf("generateSelfSignedCert returned error: %v", err)
	}
	if len(cert.Certificate) == 0 {
		t.Fatalf("expected certificate data")
	}
}
