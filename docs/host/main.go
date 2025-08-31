package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/rfwlab/rfw/docs/host/components"
	"github.com/rfwlab/rfw/v1/host"
)

func readPort() int {
	var manifest struct {
		Port int `json:"port"`
	}
	data, err := os.ReadFile("rfw.json")
	if err != nil {
		return 8080
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return 8080
	}
	if manifest.Port == 0 {
		return 8080
	}
	return manifest.Port
}

func main() {
	port := readPort()
	httpsPort := port + 1

	components.RegisterSSCHost()
	components.RegisterTwitchOAuthHost()

	go func() {
		if err := host.ListenAndServe(fmt.Sprintf(":%d", port), "client"); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	log.Fatal(host.ListenAndServeTLS(fmt.Sprintf(":%d", httpsPort), "client"))
}
