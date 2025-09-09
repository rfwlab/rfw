package devtools

import (
	"reflect"

	"github.com/rfwlab/rfw/v1/core"
)

func captureTree(c core.Component) {
	resetTree()
	walk(c, "")
}

func walk(c core.Component, parentID string) {
	id := c.GetID()
	kind := reflect.Indirect(reflect.ValueOf(c)).Type().Name()
	name := c.GetName()
	addComponent(id, kind, name, parentID)

	v := reflect.Indirect(reflect.ValueOf(c))
	depField := v.FieldByName("Dependencies")
	if !depField.IsValid() {
		if hc := v.FieldByName("HTMLComponent"); hc.IsValid() {
			depField = reflect.Indirect(hc).FieldByName("Dependencies")
		}
	}
	if depField.IsValid() {
		if deps, ok := depField.Interface().(map[string]core.Component); ok {
			for _, child := range deps {
				walk(child, id)
			}
		}
	}
}
