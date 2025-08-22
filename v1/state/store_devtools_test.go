package state

import (
	"fmt"
	"testing"
)

type capturingLogger struct{ logs []string }

func (cl *capturingLogger) Debug(format string, args ...any) {
	cl.logs = append(cl.logs, fmt.Sprintf(format, args...))
}

func TestOnChangeLoggingWithDevTools(t *testing.T) {
	// capture and restore global state
	oldLogger := logger
	cl := &capturingLogger{}
	SetLogger(cl)
	defer SetLogger(oldLogger)

	oldGSM := GlobalStoreManager
	GlobalStoreManager = &StoreManager{modules: make(map[string]map[string]*Store)}
	defer func() { GlobalStoreManager = oldGSM }()

	// store without devtools should not log
	s1 := NewStore("s1")
	s1.OnChange("a", func(any) {})
	if len(cl.logs) != 0 {
		t.Fatalf("expected no logs for store without devtools, got %v", cl.logs)
	}

	// store with devtools should log
	cl.logs = nil
	s2 := NewStore("s2", WithDevTools())
	s2.OnChange("a", func(any) {})
	if len(cl.logs) == 0 {
		t.Fatalf("expected logs for store with devtools enabled")
	}
}
