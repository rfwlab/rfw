//go:build js && wasm

package docs

import js "github.com/rfwlab/rfw/v1/js"

// LoadArticle fetches and renders the markdown document at the given path.
// It relies on the rfwLoadDoc loader injected by the docs plugin and should
// be used instead of direct js.Call invocations.
func LoadArticle(path string) {
	js.Call("rfwLoadDoc", path)
}
