//go:build js && wasm

package wasmloader

import "testing"

func TestCandidateURLs(t *testing.T) {
	tests := map[string]struct {
		url      string
		expected []string
	}{
		"wasm without query": {
			url:      "/app.wasm",
			expected: []string{"/app.wasm.br", "/app.wasm"},
		},
		"wasm with query": {
			url:      "/app.wasm?123",
			expected: []string{"/app.wasm.br?123", "/app.wasm?123"},
		},
		"already brotli": {
			url:      "/app.wasm.br",
			expected: []string{"/app.wasm.br"},
		},
		"different extension": {
			url:      "/bundle.other",
			expected: []string{"/bundle.other"},
		},
		"trimmed": {
			url:      "  /app.wasm  ",
			expected: []string{"/app.wasm.br", "/app.wasm"},
		},
		"empty": {
			url:      "   ",
			expected: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := candidateURLs(tt.url)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d urls, got %d: %v", len(tt.expected), len(got), got)
			}
			for i, val := range got {
				if val != tt.expected[i] {
					t.Fatalf("expected candidate %d to be %q, got %q", i, tt.expected[i], val)
				}
			}
		})
	}
}
