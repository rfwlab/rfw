package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

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
	var counter int
	host.Register(host.NewHostComponent("SSCHost", func(_ map[string]any) any {
		return map[string]any{"value": counter}
	}))
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			counter++
			host.Broadcast("SSCHost", map[string]any{"value": counter})
			fmt.Println("Counter:", counter)
		}
	}()
	go func() {
		if err := host.ListenAndServe(fmt.Sprintf(":%d", port), "."); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	log.Fatal(host.ListenAndServeTLS(fmt.Sprintf(":%d", httpsPort), "."))
}
