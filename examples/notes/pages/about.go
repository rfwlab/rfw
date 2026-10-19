//go:build js && wasm

package pages

import "github.com/rfwlab/rfw/v2/core"

func About() core.Component {
	c := core.NewHTMLComponent("NotesAbout", []byte(`<root><h1>About Notes</h1><p>This example uses a local Go HTTP API. Notes survive browser reloads but are lost when the API process stops.</p><a href="/">Back to notes</a></root>`), nil)
	c.Init(nil)
	return c
}
