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

func TestSignalUpdatesOnSetLang(t *testing.T) {
	trans := map[string]map[string]string{
		"en": {"hello": "Hello"},
		"it": {"hello": "Ciao"},
	}
	p := New(trans).(*Plugin)
	p.Install(nil)
	hello := Signal("hello")
	if v := hello.Get(); v != "Hello" {
		t.Fatalf("expected 'Hello', got %q", v)
	}
	SetLang("it")
	if v := hello.Get(); v != "Ciao" {
		t.Fatalf("expected 'Ciao', got %q", v)
	}
}
