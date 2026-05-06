package main

import (
	"log"

	"github.com/rfwlab/rfw/docs/host/components"
	"github.com/rfwlab/rfw/v2/host"
)

func main() {
	components.RegisterSSCHost()
	components.RegisterNetcodeHost()
	components.RegisterMultiplayerHost()
	log.Fatal(host.Start("client"))
}
