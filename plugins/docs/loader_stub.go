//go:build !js || !wasm

package docs

// LoadArticle is a no-op when not running in a js/wasm environment.
func LoadArticle(path string) {}
