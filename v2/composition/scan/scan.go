//go:build js && wasm

package scan

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rfwlab/rfw/v2/state"
	"github.com/rfwlab/rfw/v2/types"
)

type Meta struct {
	Signals       []Signal
	Stores        []Store
	Props         []Prop
	Refs          []Ref
	Hosts         []Host
	Events        []Event
	Includes      []Include
	TemplateName  string
	TemplatePath  string
	HostComponent string
}

type Signal struct{ Name string }
type Store struct{ Name string }
type Prop struct{ Name string }
type Ref struct{ Name string }
type Host struct{ Name string }
type Event struct{ Handler string }
type Include struct {
	Name  string
	Field string
}

var (
	storePtrType = reflect.TypeOf((*state.Store)(nil))
	refPtrType   = reflect.TypeOf((*types.Ref)(nil))
	viewPtrType  = reflect.TypeOf((*types.View)(nil))
)

var hostTypes = map[string]bool{
	"HInt":    true,
	"HString": true,
	"HBool":   true,
	"HFloat":  true,
	"HAny":    true,
}

func isSignalType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	if t.PkgPath() == "github.com/rfwlab/rfw/v2/state" && t.Name() == "Signal" {
		return true
	}
	_, hasGet := reflect.PtrTo(t).MethodByName("Get")
	_, hasSet := reflect.PtrTo(t).MethodByName("Set")
	_, hasRead := reflect.PtrTo(t).MethodByName("Read")
	if hasGet && hasSet && hasRead && t.NumField() >= 1 && t.Field(0).Name == "value" {
		return true
	}
	return false
}

func isSliceType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct && t.Name() == "Slice" && t.PkgPath() == "github.com/rfwlab/rfw/v2/types"
}

func isMapType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct && t.Name() == "Map" && t.PkgPath() == "github.com/rfwlab/rfw/v2/types"
}

func isPropType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct && t.Name() == "Prop" && t.PkgPath() == "github.com/rfwlab/rfw/v2/types"
}

func isHostType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	if hostTypes[t.Name()] && t.PkgPath() == "github.com/rfwlab/rfw/v2/types" {
		return true
	}
	if t.Name() == "HSlice" && t.PkgPath() == "github.com/rfwlab/rfw/v2/types" {
		return true
	}
	if t.Name() == "HMap" && t.PkgPath() == "github.com/rfwlab/rfw/v2/types" {
		return true
	}
	return false
}

var componentMethods = map[string]struct{}{
	"On":      {},
	"Prop":    {},
	"Unwrap":  {},
	"Store":   {},
	"History": {},
}

func isComponentMethod(name string) bool {
	_, ok := componentMethods[name]
	return ok
}

func Scan(v any) (*Meta, error) {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("scan: expected struct, got %v", typ)
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	m := &Meta{TemplateName: typ.Name()}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		// Skip composition.Component embedding
		if field.Anonymous && field.Type.String() == "composition.Component" {
			continue
		}

		// Check for template tag (still supported)
		if tag, ok := field.Tag.Lookup("rfw"); ok {
			if len(tag) > 9 && tag[:9] == "template:" {
				m.TemplatePath = tag[9:]
				continue
			}
			if tag == "host" {
				m.HostComponent = fieldVal.String()
				continue
			}
		}

		ft := field.Type

		switch {
		case isSignalType(ft):
			m.Signals = append(m.Signals, Signal{Name: field.Name})

		case isSliceType(ft):
			m.Signals = append(m.Signals, Signal{Name: field.Name})

		case isMapType(ft):
			m.Signals = append(m.Signals, Signal{Name: field.Name})

		case isPropType(ft):
			m.Props = append(m.Props, Prop{Name: field.Name})

		case isHostType(ft):
			m.Hosts = append(m.Hosts, Host{Name: field.Name})
			// Host fields are also signals — register for reactivity
			m.Signals = append(m.Signals, Signal{Name: field.Name})

		case ft == refPtrType:
			m.Refs = append(m.Refs, Ref{Name: field.Name})

		case ft == storePtrType:
			m.Stores = append(m.Stores, Store{Name: field.Name})

		case ft == viewPtrType:
			m.Includes = append(m.Includes, Include{
				Name:  strings.ToLower(field.Name),
				Field: field.Name,
			})

		default:
			if ft.Kind() == reflect.String {
				if tag, ok := field.Tag.Lookup("rfw"); ok && tag == "host" {
					m.HostComponent = fieldVal.String()
				}
			}
		}
	}

	// Auto-discover methods as event handlers
	ptrTyp := reflect.PtrTo(typ)
	for i := 0; i < ptrTyp.NumMethod(); i++ {
		met := ptrTyp.Method(i)
		if !met.IsExported() {
			continue
		}
		if met.Type.NumIn() != 1 || met.Type.NumOut() != 0 {
			continue
		}
		if met.Name == "OnMount" || met.Name == "OnUnmount" {
			continue
		}
		if isComponentMethod(met.Name) {
			continue
		}
		m.Events = append(m.Events, Event{Handler: met.Name})
	}

	return m, nil
}