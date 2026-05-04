package rtmleval

import (
	"testing"
)

func TestEvalOperators(t *testing.T) {
	lookup := func(name string) (any, bool) {
		m := map[string]any{
			"count": 10,
			"zero":  0,
			"name":  "world",
		}
		v, ok := m[name]
		return v, ok
	}

	tests := []struct {
		expr string
		want any
	}{
		{"count == 10", true},
		{"count != 5", true},
		{"count > 5", true},
		{"count < 20", true},
		{"count >= 10", true},
		{"count <= 10", true},
		{"count > 5 && count < 20", true},
		{"count > 5 || count < 3", true},
		{"!false", true},
		{"!(count == 5)", true},
		{"\"hello\" + \" \" + \"world\"", "hello world"},
		{"zero == 0", true},
		{"'hello' + ' ' + 'world'", "hello world"},
		{"name is 'world'", true},
		{"name is not 'world'", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := Eval(tt.expr, lookup)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			if got != tt.want {
				t.Errorf("eval(%q) = %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}
