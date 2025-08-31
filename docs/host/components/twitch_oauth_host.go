package components

import (
	"fmt"
	"os"

	helix "github.com/nicklaw5/helix/v2"
	"github.com/rfwlab/rfw/v1/host"
)

func RegisterTwitchOAuthHost() {
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
}
