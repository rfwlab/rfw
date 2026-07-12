//go:build !js

package utils

import "testing"

// The update check must stay out of the way in CI, scripts and pipes, and
// honour the explicit RFW_NO_UPDATE_CHECK opt-out.
func TestShouldSkipUpdateCheck(t *testing.T) {
	cases := []struct {
		name      string
		env       string
		stdinTTY  bool
		stdoutTTY bool
		wantSkip  bool
	}{
		{"interactive", "", true, true, false},
		{"env opt-out", "1", true, true, true},
		{"stdin piped", "", false, true, true},
		{"stdout piped", "", true, false, true},
		{"fully non-interactive", "", false, false, true},
	}
	for _, tc := range cases {
		if got := shouldSkipUpdateCheck(tc.env, tc.stdinTTY, tc.stdoutTTY); got != tc.wantSkip {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.wantSkip)
		}
	}
}
