//go:build js && wasm

package hostclient

import (
	"fmt"

	dom "github.com/rfwlab/rfw/v1/dom"
)

type domComponentRoot struct{ dom.Element }

func newComponentRoot(el dom.Element) componentRoot {
	return domComponentRoot{el}
}

func (r domComponentRoot) HostVar(name string) hostVarElement {
	selector := fmt.Sprintf(`[%s="%s"]`, hostVarAttr, name)
	return domHostVarElement{r.Element.Query(selector)}
}

func (r domComponentRoot) SetHTML(html string) {
	r.Element.SetHTML(html)
}

type domHostVarElement struct{ dom.Element }

func (e domHostVarElement) Exists() bool { return e.Value.Truthy() }

func (e domHostVarElement) Text() string { return e.Element.Text() }

func (e domHostVarElement) SetText(value string) { e.Element.SetText(value) }

func (e domHostVarElement) Attr(name string) string { return e.Element.Attr(name) }

func (e domHostVarElement) SetAttr(name, value string) { e.Element.SetAttr(name, value) }
