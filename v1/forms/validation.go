package forms

import (
	"strings"
	"unicode"

	"github.com/rfwlab/rfw/v1/state"
)

// Validator checks a field value and returns a validity flag and error code.
// When the value is valid, code should be an empty string.
type Validator func(string) (bool, string)

// Validate runs validators whenever the field signal changes and exposes the
// current validity and error code as reactive signals.
func Validate(field *state.Signal[string], validators ...Validator) (*state.Signal[bool], *state.Signal[string]) {
	valid := state.NewSignal(true)
	code := state.NewSignal("")
	state.Effect(func() func() {
		v := field.Get()
		for _, val := range validators {
			if ok, c := val(v); !ok {
				valid.Set(false)
				code.Set(c)
				return nil
			}
		}
		valid.Set(true)
		code.Set("")
		return nil
	})
	return valid, code
}

// Required ensures the string is not empty or only whitespace.
func Required(v string) (bool, string) {
	if strings.TrimSpace(v) == "" {
		return false, "IS_REQUIRED"
	}
	return true, ""
}

// Numeric ensures the value contains only digits.
func Numeric(v string) (bool, string) {
	for _, r := range v {
		if !unicode.IsDigit(r) {
			return false, "NOT_NUMERIC"
		}
	}
	return true, ""
}
