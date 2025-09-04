package main

import (
	"log"

	"github.com/rfwlab/rfw/docs/host/components"
	"github.com/rfwlab/rfw/v1/host"
)

func main() {
	components.RegisterSSCHost()
	components.RegisterTwitchOAuthHost()
	log.Fatal(host.Start("client"))
}
