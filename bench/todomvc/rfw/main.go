//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/v2/router"

	"github.com/rfwlab/bench/todomvc/pages"
)

func main() {
	router.Page("/", pages.Index)
	router.InitRouter()
	select {}
}
