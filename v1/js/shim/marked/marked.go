//go:build js && wasm

package marked

import (
	js "github.com/rfwlab/rfw/v1/js"
)

// Parse converts Markdown source to HTML using the global marked parser.
func Parse(src string) string {
	return js.Get("marked").Call("parse", src).String()
}

// Lexer tokenizes Markdown source and returns the tokens array.
func Lexer(src string) js.Value {
	return js.Get("marked").Call("lexer", src)
}
