//go:build !js || !wasm

// Package docs provides documentation loading and rendering support.
package docs

// LoadArticle is a no-op when not running in a js/wasm environment.
func LoadArticle(string) {}
