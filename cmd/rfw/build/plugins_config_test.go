//go:build !js

package build

import (
	"encoding/json"
	"testing"
)

// TestPluginsConfigDefaultsWhenMissing verifies that a missing "plugins" key
// falls back to the default set (pages), so file-based routing works out of the
// box, while an explicit block is honored verbatim.
func TestPluginsConfigDefaultsWhenMissing(t *testing.T) {
	got := pluginsConfig(nil)
	if _, ok := got["pages"]; !ok {
		t.Fatalf("nil config should enable the pages plugin by default, got %v", keys(got))
	}

	// An explicit empty block opts out of every plugin.
	empty := pluginsConfig(map[string]json.RawMessage{})
	if len(empty) != 0 {
		t.Fatalf("explicit empty config should stay empty, got %v", keys(empty))
	}

	// An explicit block is passed through unchanged.
	explicit := map[string]json.RawMessage{"tailwind": json.RawMessage("{}")}
	out := pluginsConfig(explicit)
	if _, ok := out["pages"]; ok {
		t.Fatalf("explicit config must not gain the default pages plugin, got %v", keys(out))
	}
	if _, ok := out["tailwind"]; !ok {
		t.Fatalf("explicit config should be preserved, got %v", keys(out))
	}
}

func keys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
