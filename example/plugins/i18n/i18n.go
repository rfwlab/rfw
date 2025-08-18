//go:build js && wasm

// Package i18n exposes helpers for basic string translation.
package i18n

import (
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
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
}

// New creates an i18n plugin with the supplied translation table.
func New(trans map[string]map[string]string) core.Plugin {
	return &Plugin{translations: trans, lang: "en"}
}

// Install exposes translation helpers to the JavaScript environment.
func (p *Plugin) Install(a *core.App) {
	js.Set("setLang", jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		if len(args) > 0 {
			p.lang = args[0].String()
		}
		return nil
	}))
	js.Set("t", jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		if len(args) == 0 {
			return ""
		}
		key := args[0].String()
		if val, ok := p.translations[p.lang][key]; ok {
			return val
		}
		return key
	}))
}
