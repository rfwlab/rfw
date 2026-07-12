//go:build !js

package server

import (
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// A burst of triggers within the window must coalesce into a single fire.
func TestDebouncerCoalescesBursts(t *testing.T) {
	var d debouncer
	for i := 0; i < 5; i++ {
		d.trigger(30 * time.Millisecond)
		time.Sleep(5 * time.Millisecond)
	}
	select {
	case <-d.C:
		d.fired()
	case <-time.After(time.Second):
		t.Fatalf("debouncer never fired")
	}
	if d.C != nil {
		t.Fatalf("channel not cleared after fire")
	}
	// No second fire is pending.
	d.trigger(20 * time.Millisecond)
	select {
	case <-d.C:
		d.fired()
	case <-time.After(time.Second):
		t.Fatalf("debouncer did not re-arm")
	}
}

// stopHost must terminate the child gracefully: a process that honours
// SIGTERM exits without being killed, well before the kill timeout.
func TestStopHostGraceful(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGTERM not supported on windows")
	}
	s := NewServer("0", false)
	s.hostCmd = exec.Command("sh", "-c", `trap 'exit 0' TERM; sleep 30 & wait`)
	if err := s.hostCmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	start := time.Now()
	s.stopHost()
	if elapsed := time.Since(start); elapsed >= hostStopTimeout {
		t.Fatalf("stopHost fell back to kill after %s", elapsed)
	}
	if s.hostCmd != nil {
		t.Fatalf("hostCmd not cleared")
	}
}
