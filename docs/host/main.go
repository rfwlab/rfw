package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	helix "github.com/nicklaw5/helix/v2"
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
	host.Register(host.NewHostComponent("TwitchOAuthHost", func(payload map[string]any) any {
		code, _ := payload["code"].(string)
		if code == "" {
			return nil
		}
		fmt.Println("Payload:", payload)
		client, err := helix.NewClient(&helix.Options{
			ClientID:     os.Getenv("TWITCH_CLIENT_ID"),
			ClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
			RedirectURI:  "https://localhost:8081/examples/twitch/callback",
		})
		if err != nil {
			return map[string]any{"status": err.Error()}
		}
		resp, err := client.RequestUserAccessToken(code)
		if err != nil {
			return map[string]any{"status": err.Error()}
		}
		token := resp.Data.AccessToken
		fmt.Println("Access token:", token)
		if token == "" {
			return map[string]any{"status": "missing token"}
		}
		return map[string]any{"status": "token received", "accessToken": token}
	}))
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			counter++
			host.Broadcast("SSCHost", map[string]any{"value": counter})
		}
	}()
	go func() {
		if err := host.ListenAndServe(fmt.Sprintf(":%d", port), "."); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	log.Fatal(host.ListenAndServeTLS(fmt.Sprintf(":%d", httpsPort), "."))
}
