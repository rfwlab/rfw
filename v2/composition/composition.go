//go:build js && wasm

package composition

import (
	"embed"
	"fmt"
	"io/fs"
	"reflect"

	fndi "github.com/mirkobrombin/go-foundation/pkg/di"
	"github.com/rfwlab/rfw/v2/composition/scan"
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/router"
	"github.com/rfwlab/rfw/v2/state"
	"github.com/rfwlab/rfw/v2/types"
)

var defaultContainer = fndi.New()

func Container() *fndi.Container { return defaultContainer }

var templateFS []*embed.FS

func RegisterFS(fsInstance *embed.FS) {
	templateFS = append(templateFS, fsInstance)
}

func resolveTemplatePath(path string) string {
	for _, fsInstance := range templateFS {
		if data, err := fsInstance.ReadFile(path); err == nil {
			return string(data)
		}
	}
	panic(fmt.Sprintf("composition: template %q not found in registered FS; call composition.RegisterFS() with your embed.FS", path))
}

func resolveTemplateByConvention(name string) string {
	candidates := []string{
		name + ".rtml",
		name + ".html",
	}
	for _, fsInstance := range templateFS {
		for _, c := range candidates {
			if data, err := fsInstance.ReadFile(c); err == nil {
				return string(data)
			}
		}
	}
	for _, fsInstance := range templateFS {
		var found string
		fs.WalkDir(fsInstance, ".", func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return nil
			}
			for _, c := range candidates {
				if d.Name() == c {
					if data, err := fsInstance.ReadFile(path); err == nil {
						found = string(data)
						return fs.SkipAll
					}
				}
			}
			return nil
		})
		if found != "" {
			return found
		}
	}
	return ""
}

type signalAny interface{ Read() any }

type Component struct {
	*core.HTMLComponent
	createdStores map[string]struct{}
}

func Wrap(c *core.HTMLComponent) *Component {
	comp := &Component{HTMLComponent: c, createdStores: make(map[string]struct{})}
	c.SetComponent(comp)
	return comp
}

func (c *Component) Unwrap() *core.HTMLComponent { return c.HTMLComponent }

func (c *Component) On(name string, fn func()) {
	dom.RegisterHandlerFunc(name, fn)
}

func (c *Component) Prop(key string, sig signalAny) {
	if c.HTMLComponent.Props == nil {
		c.HTMLComponent.Props = map[string]any{}
	}
	c.HTMLComponent.Props[key] = sig
}

func NewRaw(name string, tpl []byte, props map[string]any) *View {
	hc := core.NewHTMLComponent(name, tpl, props)
	defaultStore := state.GlobalStoreManager.GetStore("app", "default")
	if defaultStore == nil {
		defaultStore = state.NewStore("default", state.WithModule("app"))
	}
	hc.Init(defaultStore)
	return hc
}

func New(v any) *View {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		panic("composition.New: expected *struct")
	}
	base := typ.Elem()
	name := base.Name()
	val := reflect.ValueOf(v).Elem()

	meta, err := scan.Scan(v)
	if err != nil {
		panic(fmt.Sprintf("composition.New: scan failed: %v", err))
	}

	tpl := ""
	if meta.TemplatePath != "" {
		tpl = resolveTemplatePath(meta.TemplatePath)
	}
	if tpl == "" && meta.TemplateName != "" {
		tpl = resolveTemplateByConvention(meta.TemplateName)
	}
	if tpl == "" {
		panic("composition.New: no template found; register a convention template or use rfw:\"template:path\" tag")
	}

	hc := core.NewHTMLComponent(name, []byte(tpl), nil)
	defaultStore := state.GlobalStoreManager.GetStore("app", "default")
	if defaultStore == nil {
		defaultStore = state.NewStore("default", state.WithModule("app"))
	}
	hc.Init(defaultStore)

	comp := Wrap(hc)

	// Auto-inject router data
	for k, v := range router.RouterData() {
		if hc.Props == nil {
			hc.Props = make(map[string]any)
		}
		if _, exists := hc.Props[k]; !exists {
			hc.Props[k] = v
		}
	}

	// Wire signals: handle both value-type (t.Int) and pointer-type (*t.Int)
	for _, s := range meta.Signals {
		field := val.FieldByName(s.Name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		switch {
		case field.Kind() == reflect.Ptr && field.IsNil():
			// Nil pointer to signal: auto-init with zero value
			sig := newZeroSignalFromFieldType(field.Type())
			if sig != nil {
				field.Set(reflect.ValueOf(sig))
			}
			if sig, ok := field.Interface().(signalAny); ok {
				comp.Prop(s.Name, sig)
			}
		case field.Kind() == reflect.Ptr:
			if sig, ok := field.Interface().(signalAny); ok {
				comp.Prop(s.Name, sig)
			}
		case field.Kind() == reflect.Struct:
			if f, ok := field.Type().FieldByName("value"); ok && len(f.Index) == 1 {
				if sig, ok := field.Addr().Interface().(signalAny); ok {
					comp.Prop(s.Name, sig)
				}
			} else {
				initEmbeddedSignalPtr(field)
				if sig, ok := field.Addr().Interface().(signalAny); ok {
					comp.Prop(s.Name, sig)
				}
			}
		}
	}

	// Wire props
	for _, p := range meta.Props {
		field := val.FieldByName(p.Name)
		if !field.IsValid() {
			continue
		}
		// Props receive data from parent, for now create a placeholder signal
		if _, ok := hc.Props[p.Name]; !ok {
			sig := state.NewSignal[any](nil)
			comp.Prop(p.Name, sig)
		}
	}

	// Wire event handlers
	for _, ev := range meta.Events {
		method := val.Addr().MethodByName(ev.Handler)
		if !method.IsValid() {
			continue
		}
		fn, ok := method.Interface().(func())
		if !ok {
			continue
		}
		comp.On(ev.Handler, fn)
	}

	// Auto-discover OnMount/OnUnmount
	// Use pointer type for method lookup since methods with pointer receiver
	// (e.g. func (h *HomePage) OnMount()) are only in the pointer method set.
	ptrType := reflect.PtrTo(base)
	if m, ok := ptrType.MethodByName("OnMount"); ok {
		if m.Type.NumIn() == 1 && m.Type.NumOut() == 0 {
			method := val.Addr().MethodByName("OnMount")
			if fn, ok := method.Interface().(func()); ok {
				hc.SetOnMount(func(_ *core.HTMLComponent) { fn() })
			}
		}
	}
	if m, ok := ptrType.MethodByName("OnUnmount"); ok {
		if m.Type.NumIn() == 1 && m.Type.NumOut() == 0 {
			method := val.Addr().MethodByName("OnUnmount")
			if fn, ok := method.Interface().(func()); ok {
				hc.SetOnUnmount(func(_ *core.HTMLComponent) { fn() })
			}
		}
	}

	// Wire stores
	for _, st := range meta.Stores {
		comp.Store(st.Name)
	}

	// Wire host component (legacy string tag)
	if meta.HostComponent != "" {
		comp.HTMLComponent.AddHostComponent(meta.HostComponent)
	}

	// Wire host fields (t.HInt, t.HString, etc.) — both signal and host registration
	for _, h := range meta.Hosts {
		comp.HTMLComponent.AddHostComponent(h.Name)
	}

	// Wire includes (View dependencies)
	for _, inc := range meta.Includes {
		field := val.FieldByName(inc.Field)
		if !field.IsValid() || field.IsNil() {
			continue
		}
		if view, ok := field.Interface().(*types.View); ok {
			comp.HTMLComponent.AddDependency(inc.Name, view)
		}
	}

	// Wire Refs, register a placeholder so the DOM system can fill it later
	// (Refs are set when the component mounts and the DOM elements are available)

	

	return comp.HTMLComponent
}

func NewFrom[T any]() *View {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ == nil {
		panic("composition.NewFrom: cannot use nil type")
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		panic("composition.NewFrom: expected struct or *struct")
	}
	v := reflect.New(typ)
	return New(v.Interface())
}

func initEmbeddedSignalPtr(field reflect.Value) {
	if field.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < field.NumField(); i++ {
		f := field.Type().Field(i)
		if f.Type.Kind() != reflect.Ptr {
			continue
		}
		innerField := field.Field(i)
		if !innerField.IsNil() {
			continue
		}
		if !innerField.CanSet() {
			continue
		}
		elem := f.Type.Elem()
		if elem.Kind() != reflect.Struct {
			continue
		}
		// Check if it's *state.Signal[T] by looking for "value" field
		hasValue := false
		for j := 0; j < elem.NumField(); j++ {
			if elem.Field(j).Name == "value" {
				hasValue = true
				break
			}
		}
		if !hasValue {
			continue
		}
		// Check for signalAny interface (Get, Set, Read methods)
		_, hasGet := reflect.PtrTo(elem).MethodByName("Get")
		_, hasSet := reflect.PtrTo(elem).MethodByName("Set")
		_, hasRead := reflect.PtrTo(elem).MethodByName("Read")
		if hasGet && hasSet && hasRead {
			sig := newZeroSignalFromFieldType(f.Type)
			if sig != nil {
				innerField.Set(reflect.ValueOf(sig))
			}
		}
	}
}

func newZeroSignalFromFieldType(ptrType reflect.Type) signalAny {
	if ptrType.Kind() != reflect.Ptr {
		return nil
	}
	elem := ptrType.Elem()
	if elem.Kind() != reflect.Struct {
		return nil
	}
	// Extract T from Signal[T], look for "value" field
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		if f.Name == "value" && f.Type.Kind() != reflect.Interface {
			switch f.Type.Kind() {
			case reflect.Int:
				return state.NewSignal(int(0))
			case reflect.String:
				return state.NewSignal("")
			case reflect.Bool:
				return state.NewSignal(false)
			case reflect.Float64:
				return state.NewSignal(float64(0))
			}
			// Slice types, Signal[[]T]
			if f.Type.Kind() == reflect.Slice {
				return state.NewSignal[any](nil)
			}
			// Map types, Signal[map[K]V]
			if f.Type.Kind() == reflect.Map {
				return state.NewSignal[any](nil)
			}
		}
	}
	return state.NewSignal[any](nil)
}