//go:build js && wasm

package main

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/router"

	_ "github.com/rfwlab/examples/counter/pages"
)

func main() {
	core.SetDevMode(true)
	router.InitRouter()
	select {}
}
