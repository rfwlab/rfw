package main

import (
	"log"

	"github.com/rfwlab/rfw/v1/host"
)

func main() {
	host.Register(host.NewHostComponent("HomeHost", func(_ map[string]any) any {
		return map[string]any{"welcome": "hello from host"}
	}))
	log.Fatal(host.ListenAndServe(":8090"))
}
