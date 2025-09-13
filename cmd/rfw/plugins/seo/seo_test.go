package seo

import (
	"encoding/json"
	"os"
	"testing"
)

// TestPreAndPostBuild verifies the stub file is created and cleaned up.
func TestPreAndPostBuild(t *testing.T) {
	p := &plugin{}
	raw := json.RawMessage(`{"title":"x"}`)
	if err := p.PreBuild(raw); err != nil {
		t.Fatalf("PreBuild: %v", err)
	}
	if _, err := os.Stat("rfw_seo.go"); err != nil {
		t.Fatalf("stub not created: %v", err)
	}
	if err := p.PostBuild(nil); err != nil {
		t.Fatalf("PostBuild: %v", err)
	}
	if _, err := os.Stat("rfw_seo.go"); !os.IsNotExist(err) {
		t.Fatalf("stub not removed")
	}
}
