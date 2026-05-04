package host

import "testing"

func TestReadPortOverride(t *testing.T) {
	t.Setenv("RFW_HOST_PORT", "9095")
	if got := readPort(); got != 9095 {
		t.Fatalf("expected override port 9095, got %d", got)
	}
}
