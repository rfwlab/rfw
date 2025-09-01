package forms

import (
	"testing"

	"github.com/rfwlab/rfw/v1/state"
)

func TestRequiredValidator(t *testing.T) {
	if ok, code := Required(""); ok || code != "IS_REQUIRED" {
		t.Fatalf("empty should be invalid, got ok=%v code=%q", ok, code)
	}
	if ok, code := Required("hi"); !ok || code != "" {
		t.Fatalf("non-empty should be valid, got ok=%v code=%q", ok, code)
	}
}

func TestNumericValidator(t *testing.T) {
	if ok, code := Numeric("123"); !ok || code != "" {
		t.Fatalf("digits should be valid, got ok=%v code=%q", ok, code)
	}
	if ok, code := Numeric("abc"); ok || code != "NOT_NUMERIC" {
		t.Fatalf("letters should be invalid, got ok=%v code=%q", ok, code)
	}
}

func TestValidate(t *testing.T) {
	field := state.NewSignal("")
	valid, code := Validate(field, Required, Numeric)

	if v := valid.Get(); v {
		t.Fatalf("initial valid = %v; want false", v)
	}
	if c := code.Get(); c != "IS_REQUIRED" {
		t.Fatalf("initial code = %q; want 'IS_REQUIRED'", c)
	}

	field.Set("abc")
	if v := valid.Get(); v {
		t.Fatalf("letters should be invalid, valid=%v", v)
	}
	if c := code.Get(); c != "NOT_NUMERIC" {
		t.Fatalf("non-numeric code = %q; want 'NOT_NUMERIC'", c)
	}

	field.Set("123")
	if v := valid.Get(); !v {
		t.Fatalf("digits should be valid, valid=%v", v)
	}
	if c := code.Get(); c != "" {
		t.Fatalf("valid code = %q; want ''", c)
	}
}
