//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/forms"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/form_validation_component.rtml
var formValidationComponentTpl []byte

type FormValidationComponent struct {
	*core.HTMLComponent
	value *state.Signal[string]
	valid *state.Signal[bool]
	err   *state.Signal[string]
}

func NewFormValidationComponent() *FormValidationComponent {
	c := &FormValidationComponent{}
	c.value = state.NewSignal("")
	c.valid, c.err = forms.Validate(c.value, forms.Required, forms.Numeric)
	c.HTMLComponent = core.NewComponentWith(
		"FormValidationComponent",
		formValidationComponentTpl,
		map[string]any{
			"value": c.value,
			"valid": c.valid,
			"err":   c.err,
		},
		c,
	)
	return c
}
