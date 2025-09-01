//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/params_component.rtml
var paramsComponentTpl []byte

type ParamsComponent struct {
	*core.HTMLComponent
	ID  string
	Tab string
}

func NewParamsComponent() *ParamsComponent {
	c := &ParamsComponent{}
	c.HTMLComponent = core.NewComponentWith("ParamsComponent", paramsComponentTpl, nil, c)
	return c
}

func (p *ParamsComponent) SetRouteParams(params map[string]string) {
	p.ID = params["id"]
	p.Tab = params["tab"]
}
