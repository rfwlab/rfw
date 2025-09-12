//go:build js && wasm

// Package i18n exposes helpers for basic string translation.
package i18n

import (
	"encoding/json"

	"github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/state"
)

// Plugin installs basic internationalisation helpers. It exposes two
// JavaScript functions:
//
//	setLang(lang) - sets the active language
//	t(key)        - returns the translation for the given key
//
// Translations are provided as a map of language codes to key/value pairs.
type Plugin struct {
	translations map[string]map[string]string
	lang         string
	signals      map[string]*state.Signal[string]
}

var current *Plugin

// New creates an i18n plugin with the supplied translation table.
func New(trans map[string]map[string]string) core.Plugin {
	return &Plugin{translations: trans, lang: "en", signals: make(map[string]*state.Signal[string])}
}

// Install exposes translation helpers to the JavaScript environment.
func (p *Plugin) Install(a *core.App) {
	current = p
	js.Set("setLang", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			SetLang(args[0].String())
		}
		return nil
	}))
	js.Set("t", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return ""
		}
		key := args[0].String()
		return p.translate(key)
	}))
}

func (p *Plugin) translate(key string) string {
	if val, ok := p.translations[p.lang][key]; ok {
		return val
	}
	return key
}

func (p *Plugin) signal(key string) *state.Signal[string] {
	if sig, ok := p.signals[key]; ok {
		return sig
	}
	sig := state.NewSignal(p.translate(key))
	p.signals[key] = sig
	return sig
}

func (p *Plugin) setLang(lang string) {
	p.lang = lang
	for k, sig := range p.signals {
		sig.Set(p.translate(k))
	}
}

// SetLang updates the active language and refreshes all tracked signals.
func SetLang(lang string) {
	if current != nil {
		current.setLang(lang)
	}
}

// Signal returns a reactive translation for the given key.
func Signal(key string) *state.Signal[string] {
	if current == nil {
		return state.NewSignal("")
	}
	return current.signal(key)
}

func (p *Plugin) Build(json.RawMessage) error { return nil }
