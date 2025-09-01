//go:build js && wasm

package i18n

import "testing"

func TestNewSetsDefaultLanguage(t *testing.T) {
	trans := map[string]map[string]string{
		"en": {"hello": "Hello"},
		"it": {"hello": "Ciao"},
	}
	p := New(trans).(*Plugin)
	if p.lang != "en" {
		t.Fatalf("expected default lang 'en', got %s", p.lang)
	}
	if p.translations["it"]["hello"] != "Ciao" {
		t.Fatalf("missing translation: %v", p.translations)
	}
}
